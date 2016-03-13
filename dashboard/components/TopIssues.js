'use strict';

const React = require('react');
const { Component } = React;
const map = require('lodash.map');

const { listIssueCounts, listTopIssues } = require('../ops');

class TopIssues extends React.Component {
  constructor(props) {
    super(props);
    this.state = { status: 'loading',
                   openIssues: 0,
                   topIssues: [] };
  }

  componentWillMount() {
    const owner = window.GLOBALS.owner,
      repo = window.GLOBALS.repo;

    Promise.all([
      listIssueCounts({ owner, repo }),
      listTopIssues({ owner, repo })
    ])
    .then(([issueCounts, topIssues]) => {
      this.setState({ topIssues,
                      
                      openIssues: issueCounts[issueCounts.length-1].open_issues,
                      status: 'active' });
    });
  }

  render() {
    if (this.state.status === 'loading') {
      return (
        <div className="tile tile__top-issues">
          <div className="loader__wrapper">
            <div className="loader"></div>
          </div>
        </div>
      )
    }
    return (
      <div className="tile tile__top-issues">
        <p className="top-issues__header text-center">
          <span className="top-issues__header-icon octicon octicon-issue-opened"></span>
          <span className="top-issues__header-text"> {this.state.openIssues} Open Issues</span>
        </p>
        <div className="top-issues__list">
          {map(this.state.topIssues, (topIssue) => {
            return (
              <a className="top-issues__issue" href={topIssue.html_url} target="_blank">
                <span className="top-issues__icon octicon octicon-issue-opened"></span>
                <span className="top-issues__issue-title"> {topIssue.title}</span>
              </a>
            );
          })}
        </div>
      </div>
    )
  }
}

module.exports = TopIssues;
