'use strict';

import 'd3';
import { default as map } from 'lodash.map';
import { default as React } from 'react';

import { drawIssues } from '../helpers';
import { listIssueCounts } from '../ops';

const { Component, PropTypes } = React;

export default class PrCount extends Component {
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

PrCount.propTypes = {
  owner: PropTypes.string.isRequired,
  repo: PropTypes.string.isRequired
};
