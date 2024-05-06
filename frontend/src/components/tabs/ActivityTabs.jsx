import React from 'react';
import SiteTabs from './SiteTabs';

const ActivityTabs = () => {
    const tabsData = [
        { label: 'Comments', content: <div>Comments Go here...</div> },
        { label: 'Positions', content: <div>Positions Go here...</div> },
        { label: 'Bets', content: <div>Bets go here...</div> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;
