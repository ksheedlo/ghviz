'use strict';

const React = require('react');
const { Component } = React;

const { listStarCounts } = require('../ops');

function starCaption(stars) {
  if (stars < 10) {
    return 'Imagine the possibilities!';
  }
  if (stars < 100) {
    return 'This is starting to pick up steam!';
  }
  if (stars < 1000) {
    return 'It\'ll take over the world someday.';
  }
  if (stars < 9001) {
    return 'Look at all the Internet points!';
  }
  if (stars < 10000) {
    return 'IT\'S OVER NINE THOUSAND!';
  }
  if (stars < 100000) {
    return 'Literally bigger than jQuery.';
  }
  return 'World Domination';
}

class StarCount extends Component {
  constructor(props) {
    super(props);
    this.state = { status: 'loading', stars: 0, caption: '' };
  }

  componentWillMount() {
    const owner = window.GLOBALS.owner,
      repo = window.GLOBALS.repo;

    listStarCounts({ owner, repo }).then((starCounts) => {
      const stars = starCounts[starCounts.length-1].stars

      this.setState({ stars,
                      
                      caption: starCaption(stars),
                      status: 'active' });
    });
  }

  render() {
    if (this.state.status === 'loading') {
      return (
        <div className="tile tile__star-count">
          <div className="loader__wrapper">
            <div className="loader"></div>
          </div>
        </div>
      )
    }
    return (
      <div className="tile tile__star-count">
        <p className="star-count__text text-center">
          <span className="star-count__count">{this.state.stars} </span>
          <span className="star-count__star octicon octicon-star"></span>
        </p>
        <p className="star-count__caption text-center">{this.state.caption}</p>
      </div>
    );
  }
}

module.exports = StarCount;
