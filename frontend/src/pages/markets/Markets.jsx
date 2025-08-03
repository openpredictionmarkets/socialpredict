import React from 'react';
import SiteTabs from '../../components/tabs/SiteTabs';
import MarketsByStatusTable from '../../components/tables/MarketsByStatusTable';

function Markets() {
    const tabsData = [
        { label: 'Active', content: <MarketsByStatusTable status="active" /> },
        { label: 'Closed', content: <MarketsByStatusTable status="closed" /> },
        { label: 'Resolved', content: <MarketsByStatusTable status="resolved" /> },
        { label: 'All', content: <MarketsByStatusTable status="all" /> },
    ];

    return (
        <div className='App'>
            <div className='Center-content'>
                <div className='Center-content-header'>
                    <h1 className='text-2xl font-semibold text-gray-300 mb-6'>Markets</h1>
                </div>
                <div className='Center-content-table'>
                    <SiteTabs tabs={tabsData} />
                </div>
            </div>
        </div>
    );
}

export default Markets;
