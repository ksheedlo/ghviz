'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

/* eslint no-unused-vars: [2, { "varsIgnorePattern": "React" }] */
import { default as React } from 'react';
import { default as ReactDOM } from 'react-dom';
import Dashboard from './components/Dashboard';

ReactDOM.render(
  <Dashboard owner={window.GLOBALS.owner} repo={window.GLOBALS.repo} />,
  document.getElementById('react-dashboard')
);
