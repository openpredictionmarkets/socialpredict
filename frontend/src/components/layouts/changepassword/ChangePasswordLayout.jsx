import React, { useState, useContext } from 'react';
import { API_URL } from '../../../config';
import { useHistory } from 'react-router-dom';
import SiteButton from '../../buttons/SiteButtons';
import { RegularInput } from '../../inputs/InputBar';
import { AuthContext } from '../../../helpers/AuthContent';

function ChangePasswordLayout() {
    const [currentPassword, setCurrentPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const history = useHistory();  // Initialize useHistory hook
    const { logout } = useContext(AuthContext); // Use logout from AuthContext

    const handleCurrentPasswordChange = (event) => {
        setCurrentPassword(event.target.value);
    };

    const handleNewPasswordChange = (event) => {
        setNewPassword(event.target.value);
    };

    const handleConfirmPasswordChange = (event) => {
        setConfirmPassword(event.target.value);
    };

    const handleSubmit = async (event) => {
        event.preventDefault();
        setError('');
        setSuccess('');

        if (newPassword !== confirmPassword) {
            setError("New passwords do not match.");
            return;
        }

        try {
            const response = await fetch(`${API_URL}/api/v0/changepassword`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                },
                body: JSON.stringify({ currentPassword, newPassword })
            });
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Failed to change password');
            }
            // Set success message
            setSuccess("Password changed successfully! Logging out. Please log in with your new password.");
            // Logout user and redirect to login page after a short delay
            setTimeout(() => {
                logout();
                history.push('/login'); // Redirect to login page
            }, 2000);  // Delay of 2000 milliseconds (2 seconds)
        } catch (err) {
            console.error('Failed to change password:', err);
            setError(err.message);
        }
    };

    return (
        <div className="p-6 bg-primary-background shadow-md rounded-lg text-white">
            <h1 className="text-2xl font-bold mb-4">Change Password</h1>
            <p>Password change required. You will be automatically logged out after password has been changed.</p>
            <form onSubmit={handleSubmit} className="space-y-8">
                <RegularInput
                    type="password"
                    value={currentPassword}
                    onChange={handleCurrentPasswordChange}
                    placeholder="Current Password"
                    required
                />
                <RegularInput
                    type="password"
                    value={newPassword}
                    onChange={handleNewPasswordChange}
                    placeholder="New Password"
                    required
                />
                <RegularInput
                    type="password"
                    value={confirmPassword}
                    onChange={handleConfirmPasswordChange}
                    placeholder="Confirm New Password"
                    required
                />
                <SiteButton type="submit">
                    Save New Password
                </SiteButton>
            </form>
            {success && <p className="text-green-500">{success}</p>}
            {error && <p className="error">{error}</p>}
        </div>
    );
}

export default ChangePasswordLayout;