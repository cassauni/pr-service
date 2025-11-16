import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 50,
    duration: '30s',
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
    const prID = `k6-pr-${__VU}-${__ITER}`;

    const payload = JSON.stringify({
        pull_request_id: prID,
        pull_request_name: 'Load test PR',
        author_id: 'u001', // существует после сидирования 0002_add_data.up.sql
    });

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, {
        headers: { 'Content-Type': 'application/json' },
    });

    check(res, {
        'status is 201 or 409': (r) => r.status === 201 || r.status === 409,
    });

    sleep(0.1);
}
