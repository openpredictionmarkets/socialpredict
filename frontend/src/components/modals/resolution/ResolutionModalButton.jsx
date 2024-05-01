import React, { useState } from 'react';
import { ResolveButton, ConfirmNoButton, ConfirmYesButton } from '../../components/buttons/ResolveButtons';
import resolveMarket from './ResolutionUtils'

const ResolveModalButton = ({ marketId, token }) => {
    const [showResolveModal, setShowResolveModal] = useState(false);
    const [selectedResolution, setSelectedResolution] = useState(null);
    const [resolutionPercentage, setResolutionPercentage] = useState(0);

    const openResolveModal = () => setShowResolveModal(!showResolveModal);


    return (
        <div>
            <div className="flex-none" style={{ width: '10%' }}>
                <ResolveButton className="text-xs px-2 py-1" />
            </div>
            {showResolveModal && (
                <div className='resolve-modal p-4 bg-white shadow-lg rounded-lg'>
                    <ConfirmYesButton onClick={() => setSelectedResolution('YES')} isSelected={selectedResolution === 'YES'} />
                    <ConfirmNoButton onClick={() => setSelectedResolution('NO')} isSelected={selectedResolution === 'NO'} />
                    <input
                        type='number'
                        className='w-full p-2 border rounded'
                        value={resolutionPercentage}
                        onChange={(e) => {
                            let newValue = parseInt(e.target.value, 10);
                            setResolutionPercentage(Math.min(Math.max(newValue, 1), 99));
                        }}
                        min='1'
                        max='99'
                    />
                    <button
                        className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-700'
                        onClick={resolveMarket}
                    >
                        Confirm Resolution
                    </button>
                </div>
            )}
        </div>
    );
};

export default ResolveModalButton;
