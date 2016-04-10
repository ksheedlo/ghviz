'use strict';

import React, { PropTypes } from 'react';
import map from 'lodash.map';

export default function TopIssues({ issues, openIssues }) {
  return (
    <div>
      <p className="top-issues__header text-center">
        <span className="top-issues__header-icon octicon octicon-issue-opened"></span>
        <span className="top-issues__header-text"> {openIssues} Open Issues</span>
      </p>
      <div className="top-issues__list">
        {map(issues, (topIssue) => {
          return (
            <a className="top-issues__issue" href={topIssue.html_url} target="_blank">
              <span className="top-issues__icon octicon octicon-issue-opened"></span>
              <span className="top-issues__issue-title"> {topIssue.title}</span>
            </a>
          );
        })}
      </div>
    </div>
  );
}

TopIssues.propTypes = {
  issues: PropTypes.array.isRequired,
  openIssues: PropTypes.number.isRequired,
};
