import React from 'react';
import AdminAddUser from '../../components/layouts/admin/AddUser';
import HomeEditor from './HomeEditor';
import SiteTabs from '../../components/tabs/SiteTabs';

function AdminDashboard() {
    const tabsData = [
        { 
            label: 'Add User', 
            content: <AdminAddUser /> 
        },
        { 
            label: 'Homepage Editor', 
            content: <HomeEditor /> 
        }
    ];

    return (
        <div className="flex-col min-h-screen">
            <div className="flex-grow flex">
                <div className="flex-1">
                    <SiteTabs tabs={tabsData} />
                </div>
            </div>
        </div>
    );
}

export default AdminDashboard;
