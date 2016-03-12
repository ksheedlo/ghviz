'use strict';

const React = require('react');
const { Component } = React;
const map = require('lodash.map');

class TopPrs extends React.Component {
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
    )
  }
}

module.exports = TopPrs;