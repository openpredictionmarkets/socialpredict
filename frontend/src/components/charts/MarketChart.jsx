import React from 'react';
import CanvasJSReact from '@canvasjs/react-charts';

const CanvasJSChart = CanvasJSReact.CanvasJSChart;

const MarketChart = ({ data, title, className }) => {
    const generateDataPoints = (data) => {
    // Check if data is not undefined and is indeed an array before trying to map it
    if (data && Array.isArray(data)) {
        return data.map(item => ({
        x: new Date(item.timestamp),
        y: item.probability
        }));
    }
    return [];
    };

    const options = {
    animationEnabled: true,
    backgroundColor: "transparent",
    zoomEnabled: true,
    axisX: {
        valueFormatString: "DD MMM YY HH:mm",
        labelFontColor: "#708090",
    },
    axisY: {
        includeZero: true,
        minimum: 0,
        maximum: 1,
        labelFontColor: "#708090",
        suffix: ""
    },
    data: [{
        type: "stepArea",
        dataPoints: generateDataPoints(data)
    }]
    };

    return (
    <div className={`rounded-lg shadow p-4 ${className} overflow-hidden`}>
        <h3 className="text-lg font-medium mb-2">{title}</h3>
        <CanvasJSChart options={options} />
    </div>
    );
};

export default MarketChart;