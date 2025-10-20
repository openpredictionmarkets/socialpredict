import React, { useState } from 'react';
import CanvasJSReact from '@canvasjs/react-charts';

const CanvasJSChart = CanvasJSReact.CanvasJSChart;

const MarketChart = ({ data, currentProbability, title, className, closeDateTime, yesLabel, noLabel }) => {
  const [showInverseProbability, setShowInverseProbability] = useState(false);

  const generateDataPoints = (data, isInverse = false) => {
    let dataPoints = [];
    const now = new Date();
    const closeDate = closeDateTime ? new Date(closeDateTime) : null;
    const isMarketClosed = closeDate && closeDate < now;

    if (data && Array.isArray(data)) {
      // Filter out any probability changes that occurred after the close date for closed markets
      const filteredData = isMarketClosed 
        ? data.filter(item => new Date(item.timestamp) <= closeDate)
        : data;
      
      dataPoints = filteredData.map((item) => ({
        x: new Date(item.timestamp),
        y: isInverse ? 1 - item.probability : item.probability,
      }));
    }

    // For active markets: append current probability with current timestamp
    // For closed markets: don't extend beyond close date
    if (currentProbability !== undefined && currentProbability !== null && !isMarketClosed) {
      dataPoints.push({
        x: now,
        y: isInverse ? 1 - currentProbability : currentProbability,
      });
    }
    
    return dataPoints;
  };

  const generateChartData = () => {
    const chartData = [
      {
        type: 'stepArea',
        name: yesLabel,
        showInLegend: false, // Never show legend to prevent chart jumping
        color: showInverseProbability ? '#054A29' : '#17a2b8', // Green when showing both, blue when single
        dataPoints: generateDataPoints(data, false),
      },
    ];

    if (showInverseProbability) {
      chartData.push({
        type: 'stepArea',
        name: noLabel,
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
            className={`px-3 py-1 text-sm rounded-lg transition-colors duration-200 ${showInverseProbability
              ? 'bg-red-btn hover:bg-red-btn-hover text-white'
              : 'bg-custom-gray-light hover:bg-custom-gray-dark text-gray-300'}`}
          >
            {showInverseProbability
              ? `Show ${yesLabel} Probability`
              : `Show ${noLabel} Probability`}
          </button>
      </div>
      <CanvasJSChart options={options} />
    </div>
  );
};

export default MarketChart;
