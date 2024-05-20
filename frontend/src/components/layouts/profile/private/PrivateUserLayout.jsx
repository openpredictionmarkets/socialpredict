import React, { useState, useEffect } from 'react';
import ProfileModal from '../../../buttons/profile/ProfileModal';
import EmojiSelector from '../../../buttons/profile/EmojiSelector';
import DescriptionSelector from '../../../buttons/profile/DescriptionSelector';
import DisplayNameSelector from '../../../buttons/profile/DisplayNameSelector';

const PrivateUserInfoLayout = ({ userData }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [modalType, setModalType] = useState('');
    const [personalEmoji, setPersonalEmoji] = useState(userData ? userData.personalEmoji : '');
    const [personalDisplayName, setPersonalDisplayName] = useState(userData ? userData.displayname : '');
    const [personalDescription, setPersonalDescription] = useState(userData ? userData.description : '');

    const handleEditClick = (type) => {
        setIsModalOpen(true);
        setModalType(type);
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
    };

    const handleSaveEmoji = (emoji) => {
        setPersonalEmoji(emoji);
        handleCloseModal();
    };

    const handleSaveDisplayName = (displayname) => {
        setPersonalDisplayName(displayname);
        handleCloseModal();
    };

    const handleSaveDescription = (description) => {
        setPersonalDescription(description);
        handleCloseModal();
    };

    const renderPersonalLinks = () => {
        // Render function remains the same
    };

    return (
        <div className="overflow-auto p-6 bg-primary-background shadow-md rounded-lg">
            <h3 className="text-lg font-medium text-custom-gray-verylight mb-4">Profile Details</h3>
            <div className="divide-y divide-custom-gray-dark">
                <div className="py-4 flex justify-between items-center">
                    <span className="text-sm font-medium text-custom-gray-light">Username (Permanent):</span>
                    <span className="text-sm text-custom-gray-light">{userData.username}</span>
                </div>
                <div className="py-4 flex justify-between items-center">
                    <span className="text-sm font-medium text-custom-gray-light">Personal Emoji:</span>
                    <div className="flex items-center">
                        <span className="text-sm text-custom-gray-light mr-2">{personalEmoji}</span>
                        <button onClick={() => handleEditClick('emoji')} className="px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                            Edit
                        </button>
                    </div>
                </div>
                <div className="py-4 flex justify-between items-center">
                    <span className="text-sm font-medium text-custom-gray-light">Display Name:</span>
                    <div className="flex items-center">
                        <span className="text-sm text-custom-gray-light mr-2">{personalDisplayName}</span>
                        <button onClick={() => handleEditClick('displayname')} className="px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                            Edit
                        </button>
                    </div>
                </div>
                <div className="py-4 flex justify-between items-center">
                    <span className="text-sm font-medium text-custom-gray-light">Description:</span>
                    <div className="flex items-center">
                        <span className="text-sm text-custom-gray-light mr-2">{personalDescription}</span>
                        <button onClick={() => handleEditClick('description')} className="px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                            Edit
                        </button>
                    </div>
                </div>
                <div className="py-4 flex justify-between items-center">
                    <span className="text-sm font-medium text-custom-gray-light">Personal Links:</span>
                    <div className="flex items-center">
                        <button onClick={() => handleEditClick('links')} className="mr-2 px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                            Edit
                        </button>
                        {renderPersonalLinks()}
                    </div>
                </div>
            </div>
            {isModalOpen && (
                <ProfileModal isOpen={isModalOpen} onClose={handleCloseModal} title={`Edit ${modalType}`}>
                    {modalType === 'emoji' && <EmojiSelector onSave={handleSaveEmoji} personalfield="Emoji" />}
                    {modalType === 'displayname' && <DisplayNameSelector onSave={handleSaveDisplayName} personalfield="Display Name" />}
                    {modalType === 'description' && <DescriptionSelector onSave={handleSaveDescription} personalfield="Description" />}
                    {/* Similar conditional rendering for other selectors */}
                </ProfileModal>
            )}
        </div>
    );
};

export default PrivateUserInfoLayout;
