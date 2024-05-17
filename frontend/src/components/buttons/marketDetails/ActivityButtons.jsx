import React from 'react';
import { buttonBaseStyle } from '../BaseButton';

const ActivityModal = ({ activeTab, changeTab, bets, toggleModal }) => {
    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex justify-center items-center">
            <div className="activity-modal relative bg-blue-900 p-6 rounded-lg text-white m-6 mx-auto" style={{ width: '350px' }}>
                <div className="tabs">
                    <button onClick={() => changeTab('Comments')} className={activeTab === 'Comments' ? 'active' : ''}>Comments</button>
                    <button onClick={() => changeTab('Bets')} className={activeTab === 'Bets' ? 'active' : ''}>Bets</button>
                    <button onClick={() => changeTab('Positions')} className={activeTab === 'Positions' ? 'active' : ''}>Positions</button>
                </div>
                {activeTab === 'Bets' && (
                    <ul>
                        {bets.map(bet => <li key={bet.id}>{bet.description}</li>)}
                    </ul>
                )}
                {/* Implement similar structures for Comments and Positions */}
                <button onClick={toggleModal} className="absolute top-0 right-0 mt-4 mr-4 text-gray-400 hover:text-white">✕</button>
            </div>
        </div>
    );
};

export default ActivityModal;


