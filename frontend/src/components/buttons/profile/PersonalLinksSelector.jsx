import { API_URL } from '../../../config';
import React, { useState, useEffect } from 'react';
import SiteButton from '../SiteButtons';

const PersonalLinksSelector = ({ onSave, initialLinks }) => {
    const [links, setLinks] = useState({
        link1: initialLinks?.link1 || '',
        link2: initialLinks?.link2 || '',
        link3: initialLinks?.link3 || '',
        link4: initialLinks?.link4 || ''
    });
    const [successMessage, setSuccessMessage] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        // Update component state when initialLinks prop changes
        setLinks({
            link1: initialLinks?.link1 || '',
            link2: initialLinks?.link2 || '',
            link3: initialLinks?.link3 || '',
            link4: initialLinks?.link4 || ''
        });
    }, [initialLinks]);

    const handleSave = async () => {
        setLoading(true);
        setError('');
        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${API_URL}/api/v0/profilechange/links`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify(links),
            });
            const responseData = await response.json();
            if (response.ok) {
                console.log('Links updated successfully:', responseData);
                onSave(links); // Make sure this is passing the updated links
                setSuccessMessage('Links updated successfully.');
                setTimeout(() => setSuccessMessage(''), 3000); // Clear message after delay
            } else {
                throw new Error('Failed to update links');
            }
        } catch (error) {
            console.error('Error updating links:', error);
            setError('Failed to save links. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex flex-col items-center">
            {Object.keys(links).map((key, index) => (
                <input
                    key={key}
                    type="text"
                    value={links[key]}
                    onChange={(e) => setLinks({ ...links, [key]: e.target.value })}
                    placeholder={`Enter link ${index + 1}...`}
                    className="mb-2 px-2 py-1 border rounded text-black w-3/4"
                />
            ))}
            <SiteButton onClick={handleSave} disabled={loading} className="mt-4">
                {loading ? 'Saving...' : 'Save Links'}
            </SiteButton>
            {error && <p className="text-red-500">{error}</p>}
            {successMessage && <p className="text-green-500">{successMessage}</p>}
        </div>
    );

};

export default PersonalLinksSelector;
