'use strict';

import d3 from 'd3';
import map from 'lodash.map';
import React, { Component, PropTypes } from 'react';

const LINE_CHART_MARGIN = {
  top: 20,
  right: 20,
  bottom: 30,
  left: 50
};

export default class StarChart extends Component {
  constructor(props) {
    super(props);
  }

  componentDidMount() {
    this.props.apiClient.listStarCounts({
      owner: this.props.owner,
      repo: this.props.repo
    })
    .then((starCounts) => {
      const formattedCounts = map(starCounts, (starCount) => {
        return { stars: starCount.stars,
                 timestamp: d3.time.format.iso.parse(starCount.timestamp) };
      });

      this.refs.placeholder.removeChild(this.refs.loader);

      const starChartRect = this.refs.placeholder.getBoundingClientRect();
      const height = starChartRect.height - (LINE_CHART_MARGIN.top +
                                             LINE_CHART_MARGIN.bottom);
      const width = starChartRect.width - (LINE_CHART_MARGIN.left +
                                           LINE_CHART_MARGIN.right);

      const t = d3.time.scale()
        .range([0, width]);

      const tAxis = d3.svg.axis()
        .scale(t)
        .orient('bottom');

      const y = d3.scale.linear()
        .range([height, 0]);

      const yAxis = d3.svg.axis()
        .scale(y)
        .orient('left');

      const line = d3.svg.line()
        .x((d) => { return t(d.timestamp); })
        .y((d) => { return y(d.stars); });

      const svg = d3
        .select(this.refs.placeholder)
        .append('svg')
          .attr('class', 'chart__svg')
          .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
          .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
        .append('g')
          .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
      t.domain(d3.extent(formattedCounts, (d) => { return d.timestamp; }));
      y.domain(d3.extent(formattedCounts, (d) => { return d.stars; }));

      svg.append('g')
        .attr('class', 'chart__title')
        .append('text')
          .attr('class', 'chart__title-text')
          .attr('transform', `translate(${width / 2 - 60}, 0)`)
          .text('Stars Over Time');

      svg.append('g')
        .attr('class', 'chart__x-axis')
        .attr('transform', `translate(0,${height})`)
        .call(tAxis);

      svg.append('g')
          .attr('class', 'chart__y-axis')
          .call(yAxis)
        .append('text')
          .attr('transform', 'rotate(-90)')
          .attr('y', 6)
          .attr('dy', '.71em')
          .style('text-anchor', 'end')
          .text('Stars');

      const path = svg.append('path')
        .datum(formattedCounts)
        .attr('class', 'chart__line chart__line--orange')
        .attr('d', line);

      const totalLength = path.node().getTotalLength();

      path
        .attr('stroke-dasharray', totalLength + ' ' + totalLength)
        .attr('stroke-dashoffset', totalLength)
        .transition()
          .duration(1000)
          .ease('linear')
          .attr('stroke-dashoffset', 0);
    });
  }

  shouldComponentUpdate() {
    return false;
  }

  render() {
    return (
      <div className="tile tile__star-chart" ref="placeholder">
        <div className="loader__wrapper" ref="loader">
          <div className="loader"></div>
        </div>
      </div>
    );
  }
}

StarChart.propTypes = {
  apiClient: PropTypes.object.isRequired,
  owner: PropTypes.string.isRequired,
  repo: PropTypes.string.isRequired
};
