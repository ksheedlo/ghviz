import express, { Router } from 'express';
import fs from 'fs';
import hogan from 'hogan.js';
import httpProxy from 'http-proxy';
import { includes } from 'lodash';
import morgan from 'morgan';
import { parse } from 'url';
import process from 'process';

const app = express();
const router = Router();
const indexTpl = hogan.compile(fs.readFileSync('./index.tpl.html', 'utf-8'));
const proxy = httpProxy.createProxyServer({
  target: `${process.env.GHVIZ_API_URL}/`,
});

proxy.on('proxyReq', (proxyReq, req) => {
  proxyReq.path = parse(req.url).path.replace(/^\/gh/, '');
});

router.all('/gh/*', (req, res) => {
  proxy.web(req, res, {});
});

router.get('/', (req, res) => {
  res.send(indexTpl.render({
    owner: process.env.GHVIZ_OWNER,
    repo: process.env.GHVIZ_REPO,
  }));
});

function getConfiguredPort() {
  if (process.env.GHVIZ_WEB_PORT) {
    const integerPort = (process.env.GHVIZ_WEB_PORT|0);
    if (0 < integerPort && integerPort < 65536) {
      return integerPort;
    }
  }
  return 4001;
}

const ALLOWED_STATIC_FILES = [
  '/bundle.js',
  '/bundle.js.map',
  '/bundle.min.js',
  '/bundle.min.js.map',
  '/main.css',
  '/node_modules/bootstrap/dist/css/bootstrap.min.css',
  '/node_modules/bootstrap/dist/css/bootstrap.min.css.map',
  '/third_party/octicons/octicons.css',
  '/third_party/octicons/octicons.ttf',
  '/third_party/octicons/octicons.woff',
];

function restrictDashboardStaticFiles(req, res, next) {
  if (!includes(ALLOWED_STATIC_FILES, req.path)) {
    return res.sendStatus(404);
  }
  next();
}

app.use(morgan('short'));
app.use(router);
app.use('/dashboard', restrictDashboardStaticFiles, express.static('.'));
const configuredPort = getConfiguredPort();
app.listen(configuredPort, function () {
  /*eslint-disable no-console*/
  console.log(`${process.argv[1]} listening on port ${configuredPort}`);
  /*eslint-enable no-console*/
});
