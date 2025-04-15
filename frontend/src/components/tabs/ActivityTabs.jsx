import React from 'react';
import SiteTabs from './SiteTabs.jsx';
import BetsActivityLayout from '../layouts/activity/bets/BetsActivity.jsx';
import PositionsActivityLayout from '../layouts/activity/positions/PositionsActivity.jsx';

const ActivityTabs = ({ marketId }) => {
    const tabsData = [
        { label: 'Comments', content: <div>Comments Go here...</div> },
        { label: 'Positions', content: <PositionsActivityLayout marketId={marketId} /> },
        { label: 'Bets', content: <BetsActivityLayout marketId={marketId} /> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;