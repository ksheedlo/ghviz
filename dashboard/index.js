'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

const React = require('react');
const ReactDOM = require('react-dom');
const Dashboard = require('./components/Dashboard');

ReactDOM.render(
  <Dashboard owner={window.GLOBALS.owner} repo={window.GLOBALS.repo} />,
  document.getElementById('react-dashboard')
);
