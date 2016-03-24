'use strict';

import 'd3';
import { default as React } from 'react';
import { default as map } from 'lodash.map';

import { drawIssues } from '../helpers';
import { listIssueCounts } from '../ops';

const { Component, PropTypes } = React;

export default class IssueCount extends Component {
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
        chartLineColor: 'green',
        issueCountElement: this.refs.placeholder,
        issueCounts: formattedCounts,
        key: 'openIssues',
        loaderElement: this.refs.loader,
        title: 'Open Issues Over Time',
        yLabel: 'Open Issues'
      });
    });
  }

  shouldComponentUpdate() {
    return false;
  }

  render() {
    return (
      <div className="tile tile__issue-count" ref="placeholder">
        <div className="loader__wrapper" ref="loader">
          <div className="loader"></div>
        </div>
      </div>
    );
  }
}

IssueCount.propTypes = {
  owner: PropTypes.string.isRequired,
  repo: PropTypes.string.isRequired
};
