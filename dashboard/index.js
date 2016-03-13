'use strict';

const bluebird = require('bluebird');
if (!window.Promise) {
  window.Promise = bluebird;
}
require('whatwg-fetch');

const React = require('react');
const ReactDOM = require('react-dom');
const IssueCount = require('./components/IssueCount');
const PrCount = require('./components/PrCount');
const StarChart = require('./components/StarChart');
const StarCount = require('./components/StarCount');
const TopIssues = require('./components/TopIssues');
const TopPrs = require('./components/TopPrs');

const owner = window.GLOBALS.owner,
  repo = window.GLOBALS.repo;

ReactDOM.render(
  <IssueCount owner={owner} repo={repo} />,
  document.querySelector('.holder__issue-count')
);

ReactDOM.render(
  <PrCount owner={owner} repo={repo} />,
  document.querySelector('.holder__pr-count')
);

ReactDOM.render(
  <StarChart owner={owner} repo={repo} />,
  document.querySelector('.holder__star-chart')
);

ReactDOM.render(
  <StarCount />,
  document.querySelector('.holder__star-count')
);

ReactDOM.render(
  <TopIssues />,
  document.querySelector('.holder__top-issues')
);

ReactDOM.render(
  <TopPrs />,
  document.querySelector('.holder__top-prs')
);
