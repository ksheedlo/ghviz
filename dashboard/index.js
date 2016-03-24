'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

/* eslint no-unused-vars: [2, { "varsIgnorePattern": "React" }] */
import React from 'react';
import ReactDOM from 'react-dom';
import Dashboard from './components/Dashboard';

ReactDOM.render(
  <Dashboard owner={window.GLOBALS.owner} repo={window.GLOBALS.repo} />,
  document.getElementById('react-dashboard')
);
