import React from 'react';
import AdminAddUser from '../../components/layouts/admin/AddUser';

function AdminDashboard() {


    return (
        <div className="flex-col min-h-screen">
            <div className="flex-grow flex">
                <div className="flex-1">
                    <AdminAddUser />
                </div>
            </div>
        </div>
    );
}

export default AdminDashboard;
