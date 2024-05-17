import React, { useState } from 'react';

// Base styles for all tabs
const tabBaseStyle = "px-4 py-2 text-sm font-medium text-center cursor-pointer";
// Styles for the non-selected tabs
const tabInactiveStyle = "text-white bg-custom-gray-light border-transparent";
// Styles for the selected tab
const tabActiveStyle = "text-white bg-primary-pink";

const SiteTabs = ({ tabs }) => {
    const [activeTab, setActiveTab] = useState(tabs[0].label); // Initialize with the first tab active

    return (
        <div>
            <div className="flex border-b-2">
                {tabs.map(tab => (
                    <div
                        key={tab.label}
                        className={`${tabBaseStyle} ${activeTab === tab.label ? tabActiveStyle : tabInactiveStyle} flex-1`}
                        onClick={() => setActiveTab(tab.label)}
                    >
                        {tab.label}
                    </div>
                ))}
            </div>
            <div className="p-4">
                {tabs.map(tab => (
                    activeTab === tab.label && <div key={tab.label}>{tab.content}</div>
                ))}
            </div>
        </div>
    );
};

export default SiteTabs;
