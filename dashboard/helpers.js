'use strict';

import 'd3';

const LINE_CHART_MARGIN = {
  top: 20,
  right: 20,
  bottom: 30,
  left: 50
};

export function drawIssues({ chartLineColor,
                             issueCountElement,
                             issueCounts,
                             key,
                             loaderElement,
                             title,
                             yLabel }) {
  issueCountElement.removeChild(loaderElement);

  const boundingRect = issueCountElement.getBoundingClientRect();

  const height =
    boundingRect.height - (LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom);
  const width =
    boundingRect.width - (LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right);

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
    .y((d) => { return y(d[key]); });

  const svg = d3
    .select(issueCountElement)
    .append('svg')
      .attr('class', 'chart__svg')
      .attr('width', width + LINE_CHART_MARGIN.left + LINE_CHART_MARGIN.right)
      .attr('height', height + LINE_CHART_MARGIN.top + LINE_CHART_MARGIN.bottom)
    .append('g')
      .attr('transform', `translate(${LINE_CHART_MARGIN.left},${LINE_CHART_MARGIN.top})`);
    
  t.domain(d3.extent(issueCounts, (d) => { return d.timestamp; }));
  y.domain(d3.extent(issueCounts, (d) => { return d[key]; }));

  svg.append('g')
    .attr('class', 'chart__title')
    .append('text')
      .attr('class', 'chart__title-text')
      .attr('transform', `translate(${width / 2 - 80}, 0)`)
      .text(title);

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
      .text(yLabel);

  const path = svg.append('path')
    .datum(issueCounts)
    .attr('class', `chart__line chart__line--${chartLineColor}`)
    .attr('d', line);

  const totalLength = path.node().getTotalLength();

  path
    .attr('stroke-dasharray', totalLength + ' ' + totalLength)
    .attr('stroke-dashoffset', totalLength)
    .transition()
      .duration(1000)
      .ease('linear')
      .attr('stroke-dashoffset', 0);
};
