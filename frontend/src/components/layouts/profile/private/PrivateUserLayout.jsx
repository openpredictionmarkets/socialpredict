import React, { useState, useEffect } from 'react';
import ProfileModal from '../../../buttons/profile/ProfileModal';
import EmojiSelector from '../../../buttons/profile/EmojiSelector';
import DescriptionSelector from '../../../buttons/profile/DescriptionSelector';
import DisplayNameSelector from '../../../buttons/profile/DisplayNameSelector';
import PersonalLinksSelector from '../../../buttons/profile/PersonalLinksSelector';

const PrivateUserInfoLayout = ({ userData }) => {

    const [isModalOpen, setIsModalOpen] = useState(false);
    const [modalType, setModalType] = useState('');
    const [personalEmoji, setPersonalEmoji] = useState(userData ? userData.personalEmoji : '');
    const [personalDisplayName, setPersonalDisplayName] = useState(userData ? userData.displayname : '');
    const [personalDescription, setPersonalDescription] = useState(userData ? userData.description : '');
    const [personalLinks, setPersonalLinks] = useState({
        link1: userData?.personalLink1 || '',
        link2: userData?.personalLink2 || '',
        link3: userData?.personalLink3 || '',
        link4: userData?.personalLink4 || ''
    });

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

    const handleSavePersonalLinks = (newLinks) => {
        setPersonalLinks(newLinks);
        handleCloseModal();
    };


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

    const getModalTitle = (modalType) => {
        switch (modalType) {
            case 'emoji':
                return "Select Emoji";
            case 'displayname':
                return "Edit Display Name";
            case 'description':
                return "Edit Description";
            case 'personallinks':
                return "Edit Personal Links";
            default:
                return "";
        }
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
                    <span className="text-sm text-custom-gray-light mr-2">{renderPersonalLinks()}</span>
                    <div className="flex items-center">
                        <button onClick={() => handleEditClick('links')} className="mr-2 px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                            Edit
                        </button>
                    </div>
                </div>
            </div>
            {isModalOpen && (
                <ProfileModal isOpen={isModalOpen} onClose={handleCloseModal} title={`${getModalTitle(modalType)}`} >
                    {modalType === 'emoji' && <EmojiSelector onSave={handleSaveEmoji} />}
                    {modalType === 'displayname' && <DisplayNameSelector onSave={handleSaveDisplayName} />}
                    {modalType === 'description' && <DescriptionSelector onSave={handleSaveDescription} />}
                    {modalType === 'links' && <PersonalLinksSelector onSave={handleSavePersonalLinks} initialLinks={personalLinks} />}
                </ProfileModal>
            )}
        </div>
    );
};

export default PrivateUserInfoLayout;
