import React from 'react';
import MarketsTable from '../../components/tables/MarketTables';

function Markets() {
    return (
        <div className='App'>
        <div className='Center-content'>
            <div className='Center-content-header'>
            <h1>Markets</h1>
            </div>
            <div className='Center-content-table'>
            <MarketsTable />
            </div>
        </div>
        </div>
    );
}

export default Markets;