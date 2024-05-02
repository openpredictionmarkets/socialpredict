import React, { useState } from 'react';
import { ResolveButton, SelectNoButton, SelectYesButton, ConfirmResolveButton } from '../../buttons/ResolveButtons';
import { resolveMarket } from './ResolveUtils';

const ResolveModalButton = ({ marketId, token }) => {
    const [showResolveModal, setShowResolveModal] = useState(false);
    const [selectedResolution, setSelectedResolution] = useState(null);

    const toggleResolveModal = () => setShowResolveModal(prev => !prev);

    // handle resolution direction selection and confirmation logic
    const handleSelectNo = () => setSelectedResolution('NO');
    const handleSelectYes = () => setSelectedResolution('YES');

    const handleConfirm = () => {
        resolveMarket(marketId, token, selectedResolution)
            .then(data => {
                console.log("Resolution successful:", data);
            })
            .catch(error => {
                console.error("Failed to resolve market:", error);
            });
        setShowResolveModal(false);
    };

    return (
        <div>
            <ResolveButton onClick={toggleResolveModal} className="ml-6 w-10%" />
            {showResolveModal && (
                <div className='fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full' id="my-modal">
                    <div className='resolve-modal p-4 bg-white shadow-lg rounded-lg m-4 max-w-sm mx-auto'>
                        <SelectYesButton onClick={handleSelectYes} isSelected={selectedResolution === 'YES'} />
                        <SelectNoButton onClick={handleSelectNo} isSelected={selectedResolution === 'NO'} />
                        <ConfirmResolveButton onClick={handleConfirm} selectedResolution={selectedResolution} />
                        <button className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white" onClick={toggleResolveModal}>
                        ✕
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default ResolveModalButton;
