import React, { useRef, useState } from "react";
import * as d3 from "d3";
import Tooltip from "../utils/graphicalTools/GraphToolTips";

const MarketChartD3 = ({ data, title, className }) => {
    const d3Container = useRef(null);

    const [tooltipContent, setTooltipContent] = useState(null);
    const [tooltipPosition, setTooltipPosition] = useState({ x: 0, y: 0 });

    const drawChart = () => {
        const margin = { top: 20, right: 20, bottom: 30, left: 50 },
            width = 960 - margin.left - margin.right,
            height = 500 - margin.top - margin.bottom;

        // Append the SVG container once
        const svg = d3.select(d3Container.current)
            .append("svg")
            // Remove the explicit width and height here
            .attr('viewBox', `0 0 ${width + margin.left + margin.right} ${height + margin.top + margin.bottom}`)
            .append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`);

        const parseDate = d3.timeParse("%Y-%m-%dT%H:%M:%SZ");

        const x = d3.scaleTime().range([0, width]);
        const y = d3.scaleLinear().range([height, 0]);

        const xAxis = d3.axisBottom(x).tickFormat(d3.timeFormat("%d %b %H:%M"));
        const yAxis = d3.axisLeft(y);

        x.domain(d3.extent(data, d => parseDate(d.timestamp)));
        y.domain([0, d3.max(data, d => d.probability)]);

        // Create an area generator
        const area = d3.area()
            .x(d => x(parseDate(d.timestamp)))
            .y0(height)  // This sets the lower bound of the area (baseline)
            .y1(d => y(d.probability)) // This sets the upper bound of the area (data value)
            .curve(d3.curveStepAfter); // This makes the area step-based

        // Append the tooltip
        const tooltip = d3.select("#tooltip");

        // Append the area path using the area generator
        svg.append("path")
            .data([data])
            .attr("class", "area") // You can style this with CSS
            .attr("d", area)
            .attr("fill", "steelblue"); // This will fill the area with the steelblue color

        // Append the axes to the SVG
        svg.append("g")
            .attr("transform", `translate(0,${height})`)
            .call(xAxis);

        svg.append("g")
            .call(yAxis);

        // Create invisible circles for tooltip triggers
        svg.selectAll(".dot")
            .data(data)
            .enter().append("circle")
            .attr("class", "dot")
            .attr("cx", d => x(parseDate(d.timestamp)))
            .attr("cy", d => y(d.probability))
            .attr("r", 5) // radius of dot
            .attr("fill", "transparent") // make dot invisible
            .on("mouseover", (event, d) => {
                setTooltipContent(`Value: ${d.probability}<br/>Time: ${d.timestamp}`);
                setTooltipPosition({ x: event.pageX, y: event.pageY });
            })
            .on("mouseout", () => {
                setTooltipContent(null);
            });
        };

    React.useEffect(() => {
        if (data && d3Container.current) {
            // Check if SVG is already present, if so, remove it
            d3.select(d3Container.current).selectAll("svg").remove();
            drawChart();
        }

        // Add a resize listener to redraw the chart on window resize
        const handleResize = () => {
            // Remove the existing chart and redraw it
            d3.select(d3Container.current).selectAll("svg").remove();
            drawChart();
        };

        window.addEventListener('resize', handleResize);

        // Remove the listener on cleanup to prevent memory leaks
        return () => {
            window.removeEventListener('resize', handleResize);
        };
    }, [data]);

    return (
        <div className={`rounded-lg shadow p-4 ${className} overflow-hidden`}>
            <h3 className="text-lg font-medium mb-2">{title}</h3>
            <div ref={d3Container} />
            <Tooltip content={tooltipContent} position={tooltipPosition} />
        </div>
    );
};

export default MarketChartD3;
