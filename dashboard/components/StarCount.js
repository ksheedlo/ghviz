'use strict';

import React, { PropTypes } from 'react';

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

export default function StarCount({ count }) {
  return (
    <div>
      <p className="star-count__text text-center">
        <span className="star-count__count">{count} </span>
        <span className="star-count__star octicon octicon-star"></span>
      </p>
      <p className="star-count__caption text-center">{starCaption(count)}</p>
    </div>
  );
}

StarCount.propTypes = {
  count: PropTypes.number.isRequired,
};
