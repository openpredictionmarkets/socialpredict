import React, { useState, useEffect } from 'react';
import ActivityModal from './ActivityModal'; // Ensure this is the correct import path

const ActivityModalButton = ({ marketId, token }) => {
    const [isActivityModalOpen, setIsActivityModalOpen] = useState(false);
    const [activeTab, setActiveTab] = useState('Comments');
    const [bets, setBets] = useState([]);

    const toggleActivityModal = () => setIsActivityModalOpen(!isActivityModalOpen);
    const changeTab = (tabName) => setActiveTab(tabName);

    useEffect(() => {
        const fetchBets = async () => {
            const response = await fetch(`${API_URL}/api/v0/markets/bets/${marketId}`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            if (response.ok) {
                const data = await response.json();
                setBets(data);
            } else {
                console.error('Error fetching bets:', response.statusText);
            }
        };

        if (isActivityModalOpen && activeTab === 'Bets') {
            fetchBets();
        }
    }, [marketId, token, isActivityModalOpen, activeTab]);

    return (
        <div>
            <button onClick={toggleActivityModal} className="ml-6 w-10%">
                Activity
            </button>
            {isActivityModalOpen && (
                <ActivityModal
                    activeTab={activeTab}
                    changeTab={changeTab}
                    bets={bets}
                    toggleModal={toggleActivityModal}
                />
            )}
        </div>
    );
};

export default ActivityModalButton;
