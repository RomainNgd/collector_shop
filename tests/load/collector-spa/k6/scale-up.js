import http from 'k6/http';
import { check, sleep } from 'k6';

const baseUrl = __ENV.BASE_URL || 'http://collector.local';
const pathToHit = __ENV.PATH_TO_HIT || '/';
const hostHeader = __ENV.HOST_HEADER;
const batchSize = Number(__ENV.BATCH_SIZE || 4);
const sleepSeconds = Number(__ENV.SLEEP_SECONDS || 0.2);
const requestTimeout = __ENV.REQUEST_TIMEOUT || '5s';
const peakVUs = Number(__ENV.PEAK_VUS || 180);

export const options = {
  discardResponseBodies: true,
  scenarios: {
    scale_up_probe: {
      executor: 'ramping-vus',
      startVUs: Number(__ENV.START_VUS || 10),
      stages: [
        { duration: __ENV.WARMUP_DURATION || '30s', target: Number(__ENV.WARMUP_VUS || 40) },
        { duration: __ENV.RAMP_DURATION || '90s', target: peakVUs },
        { duration: __ENV.HOLD_DURATION || '3m', target: peakVUs },
        { duration: __ENV.COOLDOWN_DURATION || '30s', target: 0 }
      ],
      gracefulRampDown: '10s'
    }
  },
  thresholds: {
    http_req_failed: ['rate<0.10'],
    http_req_duration: ['p(95)<2500']
  }
};

function buildParams() {
  const headers = {};

  if (hostHeader) {
    headers.Host = hostHeader;
  }

  return {
    headers,
    redirects: 0,
    timeout: requestTimeout,
    tags: {
      name: 'collector-spa-home'
    }
  };
}

export default function () {
  const targetUrl = `${baseUrl}${pathToHit}`;
  const requests = [];

  for (let i = 0; i < batchSize; i += 1) {
    requests.push(['GET', `${targetUrl}?vu=${__VU}&iter=${__ITER}&req=${i}`, null, buildParams()]);
  }

  const responses = http.batch(requests);

  for (const response of responses) {
    check(response, {
      'status is 200': (res) => res.status === 200
    });
  }

  sleep(sleepSeconds);
}
