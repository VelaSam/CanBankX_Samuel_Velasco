import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import exec from 'k6/execution';

const data = new SharedArray('users', function() {
    return JSON.parse(open('./users.json')).accountIds;
});

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '10s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.01'],
    },
};

const BASE_URL = __ENV.API_URL || 'http://localhost:8080';

export function setup() {
    const loginRes = http.post(`${BASE_URL}/api/v1/identity/login`, JSON.stringify({
        email: "user1@test.com",
        password: "Password123!"
    }), { headers: { 'Content-Type': 'application/json' } });

    const mfaRes = http.post(`${BASE_URL}/api/v1/identity/login/mfa`, JSON.stringify({
        mfaToken: loginRes.json('mfaToken'),
        code: "123456"
    }), { headers: { 'Content-Type': 'application/json' } });

    return { token: mfaRes.json('accessToken') };
}

export default function (setupData) {
    const min = 0;
    const max = data.length - 1;

    // Pick two random accounts
    const sourceIdx = Math.floor(Math.random() * (max - min + 1)) + min;
    let targetIdx = Math.floor(Math.random() * (max - min + 1)) + min;
    while (targetIdx === sourceIdx) {
        targetIdx = Math.floor(Math.random() * (max - min + 1)) + min;
    }

    const sourceAccount = data[sourceIdx];
    const targetAccount = data[targetIdx];

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${setupData.token}`
        },
    };

    // 80% chance of transfer, 20% chance of history read
    if (Math.random() < 0.8) {
        const payload = JSON.stringify({
            idempotencyKey: `k6-tx-${exec.scenario.iterationInTest}-${__VU}`,
            sourceAccount: sourceAccount,
            targetAccount: targetAccount,
            amount: Math.floor(Math.random() * 50) + 1,
        });

        let res = http.post(`${BASE_URL}/api/v1/transfers/`, payload, params);
        check(res, { 'transfer status is 201 or 400 (insufficient funds)': (r) => r.status === 201 || r.status === 400 });
    } else {
        let res = http.get(`${BASE_URL}/api/v1/accounts/${sourceAccount}/history`, params);
        check(res, { 'history status is 200': (r) => r.status === 200 });
    }

    sleep(1);
}
