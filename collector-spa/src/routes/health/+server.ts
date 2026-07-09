import type { RequestHandler } from './$types';

// Endpoint de sante pour les probes Kubernetes. Il ne doit declencher ni
// rendu SSR ni appel vers go-api/Postgres: une liveness qui depend d'un
// service aval tue le front en cascade des que l'aval ralentit.
export const GET: RequestHandler = () => new Response('ok');
