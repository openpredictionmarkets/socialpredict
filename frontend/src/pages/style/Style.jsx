import React, { useState } from 'react';
import YesButton from '../../components/buttons/Buttons'; // Adjust the import path as necessary

const Style = () => {
    const [isSelected, setIsSelected] = useState(false);

    return (
    <div className="p-5">
        {/* Buttons Can Go Here */}
        <h2 className="text-2xl font-bold mb-4">Buttons</h2>
        <div className="flex flex-wrap items-center gap-4">
        <YesButton
            isSelected={isSelected}
            onClick={() => setIsSelected(!isSelected)}
        />
        </div>
    </div>
    );
};

export default Style;