import React, { useState } from 'react';
import CanvasJSReact from '@canvasjs/react-charts';

const CanvasJSChart = CanvasJSReact.CanvasJSChart;

const MarketChart = ({ data, currentProbability, title, className }) => {
  const [showInverseProbability, setShowInverseProbability] = useState(false);

  const generateDataPoints = (data, isInverse = false) => {
    let dataPoints = [];
    if (data && Array.isArray(data)) {
      dataPoints = data.map((item) => ({
        x: new Date(item.timestamp),
        y: isInverse ? 1 - item.probability : item.probability,
      }));
    }
    // Append the current probability with the current timestamp if available
    if (currentProbability !== undefined && currentProbability !== null) {
      dataPoints.push({
        x: new Date(),
        y: isInverse ? 1 - currentProbability : currentProbability,
      });
    }
    return dataPoints;
  };

  const generateChartData = () => {
    const chartData = [
      {
        type: 'stepArea',
        name: 'YES Probability',
        showInLegend: false, // Never show legend to prevent chart jumping
        color: showInverseProbability ? '#054A29' : '#17a2b8', // Green when showing both, blue when single
        dataPoints: generateDataPoints(data, false),
      },
    ];

    if (showInverseProbability) {
      chartData.push({
        type: 'stepArea',
        name: 'NO Probability',
        showInLegend: false, // Never show legend to prevent chart jumping
        color: '#D00000', // Red color for NO (using your red-btn color)
        dataPoints: generateDataPoints(data, true),
      });
    }

    return chartData;
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
      valueFormatString: '0.00',
    },
    data: generateChartData(),
  };

  return (
    <div className={`rounded-lg shadow p-4 ${className} overflow-hidden`}>
      <div className="flex justify-between items-center mb-2">
        <h3 className='text-lg font-medium'>{title}</h3>
        <button
          onClick={() => setShowInverseProbability(!showInverseProbability)}
          className={`px-3 py-1 text-sm rounded-lg transition-colors duration-200 ${
            showInverseProbability
              ? 'bg-red-btn hover:bg-red-btn-hover text-white'
              : 'bg-custom-gray-light hover:bg-custom-gray-dark text-gray-300'
          }`}
        >
          {showInverseProbability ? 'Hide NO Probability' : 'Show NO Probability'}
        </button>
      </div>
      <CanvasJSChart options={options} />
    </div>
  );
};

export default MarketChart;
