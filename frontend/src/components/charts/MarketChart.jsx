import React, { useRef } from "react";
import * as d3 from "d3";

const MarketChart = ({ data, title, className }) => {
    const d3Container = useRef(null);

    const drawChart = () => {
        const margin = { top: 20, right: 20, bottom: 30, left: 50 },
            width = 960 - margin.left - margin.right,
            height = 500 - margin.top - margin.bottom;

        const svg = d3.select(d3Container.current)
            .append("svg")
            .attr("width", width + margin.left + margin.right)
            .attr("height", height + margin.top + margin.bottom)
            .append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`);

        const parseDate = d3.timeParse("%Y-%m-%dT%H:%M:%SZ");

        const x = d3.scaleTime().range([0, width]);
        const y = d3.scaleLinear().range([height, 0]);

        const xAxis = d3.axisBottom(x).tickFormat(d3.timeFormat("%d %b"));
        const yAxis = d3.axisLeft(y);

        x.domain(d3.extent(data, d => parseDate(d.timestamp)));
        y.domain([0, d3.max(data, d => d.probability)]);

        const line = d3.line()
            .x(d => x(parseDate(d.timestamp)))
            .y(d => y(d.probability))
            .curve(d3.curveStepAfter);

        svg.append("path")
            .data([data])
            .attr("class", "line")
            .attr("d", line)
            .attr("fill", "none")
            .attr("stroke", "steelblue");

        svg.append("g")
            .attr("transform", `translate(0,${height})`)
            .call(xAxis);

        svg.append("g")
            .call(yAxis);
    };

    // When the component mounts, draw the chart:
    React.useEffect(() => {
        if (data && d3Container.current) {
            console.log("Data received by MarketChart", data);
            // Check if SVG is already present, if so, remove it
            d3.select(d3Container.current).selectAll("svg").remove();
            drawChart();
        }
    }, []); // Empty dependency array means it will only run once on mount

    return (
        <div className={`rounded-lg shadow p-4 ${className} overflow-hidden`}>
            <h3 className="text-lg font-medium mb-2">{title}</h3>
            <div ref={d3Container} />
        </div>
    );
};

export default MarketChart;
