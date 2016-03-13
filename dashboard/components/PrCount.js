'use strict';

const React = require('react');
const { Component } = React;

const d3 = require('d3'),
  map = require('lodash.map');

const { drawIssues } = require('../helpers');
const { listIssueCounts } = require('../ops');

class PrCount extends Component {
  constructor(props) {
    super(props);
  }

  componentDidMount() {
    listIssueCounts({
      owner: this.props.owner,
      repo: this.props.repo
    })
    .then((issueCounts) => {
      const formattedCounts = map(issueCounts, (issueCount) => {
        return { openIssues: issueCount.open_issues,
                 openPrs: issueCount.open_prs,
                 timestamp: d3.time.format.iso.parse(issueCount.timestamp) };
      });

      drawIssues({
        chartLineColor: 'blue',
        issueCountElement: this.refs.placeholder,
        issueCounts: formattedCounts,
        key: 'openPrs',
        loaderElement: this.refs.loader,
        title: 'Open PRs Over Time',
        yLabel: 'Open PRs'
      });
    });
  }

  shouldComponentUpdate() {
    return false;
  }

  render() {
    return (
      <div className="tile tile__pr-count" ref="placeholder">
        <div className="loader__wrapper" ref="loader">
          <div className="loader"></div>
        </div>
      </div>
    );
  }
}

module.exports = PrCount;
