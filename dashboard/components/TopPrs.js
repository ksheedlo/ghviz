'use strict';

const React = require('react');
const { Component } = React;
const map = require('lodash.map');

const { listIssueCounts, listTopPrs } = require('../ops');

class TopPrs extends React.Component {
  constructor(props) {
    super(props);
    this.state = { status: 'loading',
                   openPrs: 0,
                   topPrs: [] };
  }

  componentWillMount() {
    const owner = window.GLOBALS.owner,
      repo = window.GLOBALS.repo;

    Promise.all([
      listIssueCounts({ owner, repo }),
      listTopPrs({ owner, repo })
    ])
    .then(([issueCounts, topPrs]) => {
      this.setState({ status: 'active',
                      openPrs: issueCounts[issueCounts.length-1].open_prs,
                      topPrs: topPrs });
    });
  }

  render() {
    if (this.state.status === 'loading') {
      return (
        <div className="tile tile__top-prs">
          <div className="loader__wrapper">
            <div className="loader"></div>
          </div>
        </div>
      )
    }
    return (
      <div className="tile tile__top-prs">
        <p className="top-issues__header text-center">
          <span className="top-issues__header-icon octicon octicon-git-pull-request"></span>
          <span className="top-issues__header-text"> {this.state.openPrs} Open PRs</span>
        </p>
        <div className="top-issues__list">
          {map(this.state.topPrs, (topPr) => {
            return (
              <a className="top-issues__issue" href={topPr.html_url} target="_blank">
                <span className="top-issues__icon octicon octicon-git-pull-request"></span>
                <span className="top-issues__issue-title"> {topPr.title}</span>
              </a>
            );
          })}
        </div>
      </div>
    )
  }
}

module.exports = TopPrs;
