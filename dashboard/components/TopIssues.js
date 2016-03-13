'use strict';

const React = require('react');
const { Component, PropTypes } = React;
const map = require('lodash.map');

class TopIssues extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div>
        <p className="top-issues__header text-center">
          <span className="top-issues__header-icon octicon octicon-issue-opened"></span>
          <span className="top-issues__header-text"> {this.props.openIssues} Open Issues</span>
        </p>
        <div className="top-issues__list">
          {map(this.props.issues, (topIssue) => {
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
}

TopIssues.propTypes = {
  issues: PropTypes.array.isRequired,
  openIssues: PropTypes.number.isRequired
};

module.exports = TopIssues;
