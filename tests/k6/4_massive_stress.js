import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 4000 }, 
        { duration: '3m', target: 4000 }, 
        { duration: '1m', target: 0 },   
    ],
    thresholds: {
        http_req_duration: ['p(95)<5000'],
    },
};

export default function () {
    const BASE_URL = __ENV.API_URL || 'http://localhost:8080';
    
    let res = http.get(`${BASE_URL}/health`);
    check(res, { 'status is 200': (r) => r.status === 200 });

    sleep(1);
}
