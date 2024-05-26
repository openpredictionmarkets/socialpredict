import React from 'react';
import ChangePasswordLayout from '../../components/layouts/changepassword/ChangePasswordLayout';

function ChangePassword() {


    return (
        <div className="flex-col min-h-screen">
            <div className="flex-grow flex">
                <div className="flex-1">
                    <ChangePasswordLayout />
                </div>
            </div>
        </div>
    );
}

export default ChangePassword;
