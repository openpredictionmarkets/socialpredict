import React from 'react';
import ModeratorMarketReview from '../../components/layouts/admin/ModeratorMarketReview';
import UserQueue from '../../components/layouts/admin/UserQueue';
import CmsDashboard from './CmsDashboard';
import SiteTabs from '../../components/tabs/SiteTabs';

function AdminDashboard({
    defaultTab = 'Review Markets',
    defaultReviewTab = 'Pending Review',
    defaultUserTab = 'Non-Moderators',
}) {
    const tabsData = [
        {
            label: 'Review Markets',
            content: <ModeratorMarketReview defaultTab={defaultReviewTab} />
        },
        {
            label: 'User Governance',
            content: <UserQueue defaultTab={defaultUserTab} />
        },
        {
            label: 'CMS',
            content: <CmsDashboard />
        }
    ];

    return (
        <div className="flex-col min-h-screen">
            <div className="flex-grow flex">
                <div className="flex-1">
                    <SiteTabs tabs={tabsData} defaultTab={defaultTab} />
                </div>
            </div>
        </div>
    );
}

export default AdminDashboard;
