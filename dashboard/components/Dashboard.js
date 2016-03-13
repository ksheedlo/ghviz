'use strict';

const React = require('react');
const { Component, PropTypes } = React;

const IssueCount = require('./IssueCount');
const PrCount = require('./PrCount');
const StarChart = require('./StarChart');
const StarCount = require('./StarCount');
const TopIssues = require('./TopIssues');
const TopPrs = require('./TopPrs');

const { listStarCounts,
        listTopPrs,
        listTopIssues,
        listIssueCounts } = require('../ops');

class Dashboard extends Component {
  constructor(props) {
    super(props);
    this.state = { loadingTopPrs: true,
                   loadingStarCount: true,
                   loadingTopIssues: true,
                   openPrs: 0,
                   openIssues: 0,
                   starCount: 0,
                   topPrs: null,
                   topIssues: null };
  }

  componentWillMount() {
    const queryProps = { owner: this.props.owner, repo: this.props.repo };

    listStarCounts(queryProps).then((starCounts) => {
      this.setState({ ...this.state,

                      loadingStarCount: false,
                      starCount: starCounts[starCounts.length-1].stars });
    });

    const issueCountsPromise = listIssueCounts(queryProps);

    Promise.all([
      issueCountsPromise,
      listTopIssues(queryProps)
    ]).then(([issueCounts, topIssues]) => {
      this.setState({ ...this.state,

                      topIssues,

                      openIssues: issueCounts[issueCounts.length-1].open_issues,
                      loadingTopIssues: false });
    });

    Promise.all([
      issueCountsPromise,
      listTopPrs(queryProps)
    ]).then(([issueCounts, topPrs]) => {
      this.setState({ ...this.state,

                      topPrs,

                      loadingTopPrs: false,
                      openPrs: issueCounts[issueCounts.length-1].open_prs });
    });
  }

  renderLoader() {
    return (
      <div className="loader__wrapper">
        <div className="loader"/>
      </div>
    );
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
            <IssueCount owner={this.props.owner} repo={this.props.repo} />
          </div>
          <div className="col-sm-4">
            <div className="tile">
              {this.renderTopPrs()}
            </div>
          </div>
        </div>
        <div className="row">
          <div className="col-sm-4">
            <StarChart owner={this.props.owner} repo={this.props.repo} />
          </div>
          <div className="col-sm-4">
            <div className="tile">
              {this.renderTopIssues()}
            </div>
          </div>
          <div className="col-sm-4">
            <PrCount owner={this.props.owner} repo={this.props.repo} />
          </div>
        </div>
      </div>
    );
  }
}

Dashboard.propTypes = {
  owner: PropTypes.string.isRequired,
  repo: PropTypes.string.isRequired
};

module.exports = Dashboard;
