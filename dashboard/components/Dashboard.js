'use strict';

import React, { Component, PropTypes } from 'react';
import IssueCount from './IssueCount';
import PrCount from './PrCount';
import StarChart from './StarChart';
import StarCount from './StarCount';
import TopContributors from './TopContributors';
import TopIssues from './TopIssues';
import TopPrs from './TopPrs';

export default class Dashboard extends Component {
  constructor(props) {
    super(props);
    this.state = { loadingTopPrs: true,
                   loadingStarCount: true,
                   loadingTopIssues: true,
                   loadingTopContributors: true,
                   openPrs: 0,
                   openIssues: 0,
                   starCount: 0,
                   topContributors: null,
                   topPrs: null,
                   topIssues: null };
  }

  componentWillMount() {
    const queryProps = { owner: this.props.owner, repo: this.props.repo };

    this.props.apiClient.listStarCounts(queryProps).then((starCounts) => {
      this.setState({ ...this.state,

                      loadingStarCount: false,
                      starCount: starCounts[starCounts.length-1].stars });
    });

    const issueCountsPromise = this.props.apiClient.listIssueCounts(queryProps);

    Promise.all([
      issueCountsPromise,
      this.props.apiClient.listTopIssues(queryProps)
    ]).then(([issueCounts, topIssues]) => {
      this.setState({ ...this.state,

                      topIssues,

                      openIssues: issueCounts[issueCounts.length-1].open_issues,
                      loadingTopIssues: false });
    });

    Promise.all([
      issueCountsPromise,
      this.props.apiClient.listTopPrs(queryProps)
    ]).then(([issueCounts, topPrs]) => {
      this.setState({ ...this.state,

                      topPrs,

                      loadingTopPrs: false,
                      openPrs: issueCounts[issueCounts.length-1].open_prs });
    });

    this.props.apiClient.listTopContributors({
      ...queryProps,

      date: new Date()
    }).then((topContributors) => {
      this.setState({ ...this.state,

                      topContributors,

                      loadingTopContributors: false });
    });
  }

  renderLoader() {
    return (
      <div className="loader__wrapper">
        <div className="loader"/>
      </div>
    );
  }

  renderHighScores() {
    if (this.state.loadingTopContributors) {
      return this.renderLoader();
    }
    return <TopContributors contributors={this.state.topContributors} />;
  }

  renderStarCount() {
    if (this.state.loadingStarCount) {
      return this.renderLoader();
    }
    return <StarCount count={this.state.starCount} />;
  }

  renderTopPrs() {
    if (this.state.loadingTopPrs) {
      return this.renderLoader();
    }
    return <TopPrs prs={this.state.topPrs} openPrs={this.state.openPrs} />;
  }

  renderTopIssues() {
    if (this.state.loadingTopIssues) {
      return this.renderLoader();
    }
    return <TopIssues issues={this.state.topIssues} openIssues={this.state.openIssues} />;
  }

  render() {
    return (
      <div>
        <div className="row">
          <div className="col-sm-4">
            <div className="tile">
              {this.renderStarCount()}
            </div>
          </div>
          <div className="col-sm-4">
            <IssueCount apiClient={this.props.apiClient}
              owner={this.props.owner}
              repo={this.props.repo} />
          </div>
          <div className="col-sm-4">
            <div className="tile">
              {this.renderTopPrs()}
            </div>
          </div>
        </div>
        <div className="row">
          <div className="col-sm-4">
            <StarChart apiClient={this.props.apiClient}
              owner={this.props.owner}
              repo={this.props.repo} />
          </div>
          <div className="col-sm-4">
            <div className="tile">
              {this.renderTopIssues()}
            </div>
          </div>
          <div className="col-sm-4">
            <PrCount apiClient={this.props.apiClient}
              owner={this.props.owner}
              repo={this.props.repo} />
          </div>
        </div>
        <div className="row">
          <div className="col-sm-4">
            <div className="tile">
              {this.renderHighScores()}
            </div>
          </div>
        </div>
      </div>
    );
  }
}

Dashboard.propTypes = {
  apiClient: PropTypes.object.isRequired,
  owner: PropTypes.string.isRequired,
  repo: PropTypes.string.isRequired
};
