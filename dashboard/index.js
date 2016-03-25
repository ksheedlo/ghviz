'use strict';

import bluebird from 'bluebird';
if (!window.Promise) {
  window.Promise = bluebird;
}
import 'whatwg-fetch';

/* eslint no-unused-vars: [2, { "varsIgnorePattern": "React" }] */
import React from 'react';
import ReactDOM from 'react-dom';
import Dashboard from './components/Dashboard';

ReactDOM.render(
  <Dashboard owner={window.GLOBALS.owner} repo={window.GLOBALS.repo} />,
  document.getElementById('react-dashboard')
);
