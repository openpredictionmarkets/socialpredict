import React from 'react';

const PublicUserInfoLayout = ({ userData }) => {
    if (!userData) {
        return <div>Loading...</div>;
    }

    const renderPersonalLinks = () => {
        const linkKeys = ['personalink1', 'personalink2', 'personalink3', 'personalink4'];
        return linkKeys.map(key => {
            const link = userData[key];
            return link ? (
                <div key={key} className='nav-link text-info-blue hover:text-blue-800'>
                    <a
                        href={link}
                        target='_blank'
                        rel='noopener noreferrer'
                    >
                        {link}
                    </a>
                </div>
            ) : null;
        });
    };

    return (
        <div className="p-6 bg-primary-background shadow-md rounded-lg">
            <h3 className="text-lg font-medium text-custom-gray-verylight mb-4">User Info</h3>
            <table className="min-w-full divide-y divide-custom-gray-dark">
                <tbody className="bg-primary-background divide-y divide-custom-gray-dark">
                    <tr>
                        <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Username (Permanent):</td>
                        <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.username}</td>
                    </tr>
                    <tr>
                        <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Personal Emoji:</td>
                        <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.personalEmoji}</td>
                    </tr>
                    <tr>
                        <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Display Name:</td>
                        <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.displayname}</td>
                    </tr>
                    <tr>
                        <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Description:</td>
                        <td className="px-6 py-4 text-sm text-custom-gray-light">{userData.description}</td>
                    </tr>
                    <tr>
                        <td className="px-6 py-4 text-sm font-medium text-custom-gray-light">Personal Links:</td>
                        <td className="px-6 py-4 text-sm text-custom-gray-light">{renderPersonalLinks()}</td>
                    </tr>
                </tbody>
            </table>
        </div>
    );
};

export default PublicUserInfoLayout;
