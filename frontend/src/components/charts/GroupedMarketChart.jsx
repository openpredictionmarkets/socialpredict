import React from 'react';
import CanvasJSReact from '@canvasjs/react-charts';

const CanvasJSChart = CanvasJSReact.CanvasJSChart;

const SERIES_COLORS = [
  '#38bdf8',
  '#f97316',
  '#22c55e',
  '#e879f9',
  '#facc15',
  '#a78bfa',
  '#14b8a6',
  '#fb7185',
];

const toNumber = (value, fallback = 0.5) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
};

const toDate = (value) => {
  const date = value ? new Date(value) : null;
  return date && !Number.isNaN(date.getTime()) ? date : null;
};

const answerLabel = (answer, index) => (
  answer?.answerLabel || answer?.market?.market?.marketGroup?.answerLabel || `Answer ${index + 1}`
);

const answerProbability = (answer) => toNumber(
  answer?.probabilityChanges?.[answer.probabilityChanges.length - 1]?.probability
    ?? answer?.summary?.lastProbability
    ?? answer?.market?.lastProbability
    ?? answer?.market?.market?.initialProbability,
  0.5,
);

const answerPoints = (answer) => {
  const changes = Array.isArray(answer?.probabilityChanges)
    ? answer.probabilityChanges
    : Array.isArray(answer?.summary?.probabilityChanges)
      ? answer.summary.probabilityChanges
      : [];
  const now = new Date();
  const closeDate = toDate(answer?.market?.market?.resolutionDateTime);
  const isClosed = closeDate && closeDate <= now;
  const filteredChanges = isClosed
    ? changes.filter((point) => toDate(point.timestamp)?.getTime() <= closeDate.getTime())
    : changes;
  const points = filteredChanges
    .map((point) => ({
      x: toDate(point.timestamp),
      y: toNumber(point.probability, answerProbability(answer)),
    }))
    .filter((point) => point.x);

  const finalX = isClosed ? closeDate : now;
  const finalY = points[points.length - 1]?.y ?? answerProbability(answer);
  if (finalX && (!points.length || points[points.length - 1].x.getTime() < finalX.getTime())) {
    points.push({ x: finalX, y: finalY });
  }
  return points.length > 0 ? points : [{ x: now, y: answerProbability(answer) }];
};

export default function GroupedMarketChart({ answers = [], title = 'Answer Probabilities' }) {
  const sortedAnswers = [...answers].sort((left, right) => (
    Number(left.displayOrder || 0) - Number(right.displayOrder || 0)
  ));

  const options = {
    animationEnabled: true,
    backgroundColor: 'transparent',
    zoomEnabled: true,
    legend: {
      fontColor: '#cbd5e1',
      cursor: 'pointer',
    },
    axisX: {
      valueFormatString: 'DD MMM YY HH:mm',
      labelFontColor: '#94a3b8',
    },
    axisY: {
      includeZero: true,
      minimum: 0,
      maximum: 1,
      labelFontColor: '#94a3b8',
      valueFormatString: '0.00',
    },
    toolTip: {
      shared: true,
    },
    data: sortedAnswers.map((answer, index) => ({
      type: 'stepLine',
      name: answerLabel(answer, index),
      showInLegend: true,
      color: SERIES_COLORS[index % SERIES_COLORS.length],
      markerSize: 4,
      dataPoints: answerPoints(answer),
    })),
  };

  return (
    <div className='rounded-lg bg-gray-950/40 p-4 shadow'>
      <h3 className='mb-3 text-lg font-medium text-white'>{title}</h3>
      <CanvasJSChart options={options} />
    </div>
  );
}
