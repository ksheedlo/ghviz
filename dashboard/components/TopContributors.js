'use strict';

import { default as React } from 'react';
import { default as map } from 'lodash.map';

const { Component, PropTypes } = React;

const PLACES = ['1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th'];
const BANGS = ['!!!', '!!', '!', '', '', '', '', ''];

export default class TopContributors extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <div className="top-contributors">
        <p className="top-contributors__header text-center">
          Top Contributors
        </p>
        <div className="top-contributors__list">
          {map(this.props.contributors, (contributor, i) => {
            return (
              <a className="top-contributors__contributor" target="_blank"
                href={'https://github.com/' + contributor.actor_id}>
                <div className="row">
                  <div className="col-xs-2">{PLACES[i]}</div>
                  <div className="col-xs-6">{contributor.actor_id}</div>
                  <div className="col-xs-2">{contributor.score}</div>
                  <div className="col-xs-2">{BANGS[i]}</div>
                </div>
              </a>
            );
          })}
        </div>
      </div>
    );
  }
}

TopContributors.propTypes = {
  contributors: PropTypes.array.isRequired
};
