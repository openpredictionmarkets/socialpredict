import React from 'react';
import CanvasJSReact from 'react-charts';

const CanvasJSChart = CanvasJSReact.CanvasJSChart;

const MarketChart = ({ data, currentProbability, title, className }) => {
  const generateDataPoints = (data) => {
    let dataPoints = [];
    if (data && Array.isArray(data)) {
      dataPoints = data.map((item) => ({
        x: new Date(item.timestamp),
        y: item.probability,
      }));
    }
    // Append the current probability with the current timestamp if available
    if (currentProbability !== undefined && currentProbability !== null) {
      dataPoints.push({
        x: new Date(),
        y: currentProbability,
      });
    }
    return dataPoints;
  };

  const options = {
    animationEnabled: true,
    backgroundColor: 'transparent',
    zoomEnabled: true,
    axisX: {
      valueFormatString: 'DD MMM YY HH:mm',
      labelFontColor: '#708090',
    },
    axisY: {
      includeZero: true,
      minimum: 0,
      maximum: 1,
      labelFontColor: '#708090',
      suffix: '',
    },
    data: [
      {
        type: 'stepArea',
        dataPoints: generateDataPoints(data),
      },
    ],
  };

  return (
    <div className={`rounded-lg shadow p-4 ${className} overflow-hidden`}>
      <h3 className='text-lg font-medium mb-2'>{title}</h3>
      <CanvasJSChart options={options} />
    </div>
  );
};

export default MarketChart;
