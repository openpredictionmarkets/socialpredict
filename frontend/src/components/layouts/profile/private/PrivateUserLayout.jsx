import React, { useState } from 'react';
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
        personallink1: userData?.personalLink1 || '',
        personallink2: userData?.personalLink2 || '',
        personallink3: userData?.personalLink3 || '',
        personallink4: userData?.personalLink4 || ''
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
        setPersonalLinks(prevLinks => ({
            ...prevLinks,
            ...newLinks
        }));
        handleCloseModal();
    };

    //const renderPersonalLinks = () => {
    //    const linkKeys = ['personalink1', 'personalink2', 'personalink3', 'personalink4'];
    //    return linkKeys.map(key => {
    //        const link = userData[key];
    //        return link ? (
    //            <div key={key} className='nav-link text-info-blue hover:text-blue-800'>
    //                <a
    //                    href={link}
    //                    target='_blank'
    //                    rel='noopener noreferrer'
    //                >
    //                    {link}
    //                </a>
    //            </div>
    //        ) : null;
    //    });
    //};


    const renderPersonalLinks = () => {
        // Match the keys with the state keys used in `setLinks`
        const linkKeys = ['personalLink1', 'personalLink2', 'personalLink3', 'personalLink4'];
        return linkKeys.map(key => {
            const link = personalLinks[key];  // Use the current state to render links
            return link ? (
                <div key={key} className='nav-link text-info-blue hover:text-blue-800'>
                    <a href={link} target='_blank' rel='noopener noreferrer'>
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
                {[
                    { label: 'Username (Permanent)', value: userData.username },
                    { label: 'Personal Emoji', value: personalEmoji, type: 'emoji' },
                    { label: 'Display Name', value: personalDisplayName, type: 'displayname' },
                    { label: 'Description', value: personalDescription, type: 'description' },
                    { label: 'Personal Links', value: renderPersonalLinks(), type: 'links' }
                ].map(item => (
                    <div key={item.label} className="py-4 grid grid-cols-3 items-center">
                        <span className="text-sm font-medium text-custom-gray-light">{item.label}:</span>
                        <span className="text-sm text-custom-gray-light col-span-1">{item.value}</span>
                        {item.type && (
                            <button onClick={() => handleEditClick(item.type)} className="justify-self-end px-2 py-1 bg-blue-500 hover:bg-blue-700 text-white rounded">
                                Edit
                            </button>
                        )}
                    </div>
                ))}
            </div>
            {isModalOpen && (
                <ProfileModal isOpen={isModalOpen} onClose={handleCloseModal} title={`${getModalTitle(modalType)}`}>
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
