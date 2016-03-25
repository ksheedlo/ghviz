'use strict';

import bluebird from 'bluebird';
if (!window.Promise) {
  window.Promise = bluebird;
}
import 'whatwg-fetch';

/* eslint no-unused-vars: [2, { "varsIgnorePattern": "React" }] */
import React from 'react';
import ReactDOM from 'react-dom';
import ApiClient from './api-client';
import Cache from './cache';
import Dashboard from './components/Dashboard';

const apiClient = new ApiClient({ cache: new Cache({ maxAge: 1000 * 60 * 4 }) });

ReactDOM.render(
  <Dashboard apiClient={apiClient}
    owner={window.GLOBALS.owner}
    repo={window.GLOBALS.repo} />,
  document.getElementById('react-dashboard')
);
