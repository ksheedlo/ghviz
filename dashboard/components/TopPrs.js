'use strict';

import React, { Component, PropTypes } from 'react';
import map from 'lodash.map';

export default class TopPrs extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div>
        <p className="top-issues__header text-center">
          <span className="top-issues__header-icon octicon octicon-git-pull-request"></span>
          <span className="top-issues__header-text"> {this.props.openPrs} Open PRs</span>
        </p>
        <div className="top-issues__list">
          {map(this.props.prs, (topPr) => {
            return (
              <a className="top-issues__issue" href={topPr.html_url} target="_blank">
                <span className="top-issues__icon octicon octicon-git-pull-request"></span>
                <span className="top-issues__issue-title"> {topPr.title}</span>
              </a>
            );
          })}
        </div>
      </div>
    );
  }
}

TopPrs.propTypes = {
  openPrs: PropTypes.number.isRequired,
  prs: PropTypes.array.isRequired,
};
