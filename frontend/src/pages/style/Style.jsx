import React, { useState } from 'react';
import {BetYesButton, BetNoButton, ResolveButton} from '../../components/buttons/Buttons'; // Adjust the import path as necessary
import Sidebar from '../../components/sidebar/Sidebar';

const Style = () => {
    const [isSelected, setIsSelected] = useState(false);

    return (
    <div className="overflow-x-auto">
    <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
        <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Component
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Description
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Import
            </th>
        </tr>
        </thead>
        <tbody className="bg-white divide-y divide-gray-200">
        <tr>
            <td className="px-6 py-4 whitespace-nowrap">
                {/* This div might be used to contain a static image or a mock-up if you're not embedding the actual Sidebar */}
                <div className="flex items-center">
                    {/* If you decide to include the actual Sidebar component, ensure it's visually and functionally compatible within this constrained space */}
                    <Sidebar />
                </div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                Sidebar
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">
                <code>{`import Sidebar from '../../components/sidebar/Sidebar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4 whitespace-nowrap">
            <div className="flex flex-wrap items-center gap-4">
                <BetYesButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            Bet YES Button
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">
            <code>{`import BetYesButton from '../../components/buttons/Buttons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4 whitespace-nowrap">
            <div className="flex flex-wrap items-center gap-4">
                <BetNoButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            Bet NO Button
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">
            <code>{`import BetNoButton from '../../components/buttons/Buttons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4 whitespace-nowrap">
            <div className="flex flex-wrap items-center gap-4">
                <ResolveButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            Neutral Button
            </td>
            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-500">
            <code>{`import NeutralButton from '../../components/buttons/Buttons';`}</code>
            </td>
        </tr>
        </tbody>
    </table>
    </div>

    );
};

export default Style;