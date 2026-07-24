/** @vitest-environment jsdom */
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ProjectionErrorDetails } from './SellSharesLayout';

describe('ProjectionErrorDetails', () => {
    it('renders projection values as stacked rows so long labels do not overlap in narrow dialogs', () => {
        render(
            <ProjectionErrorDetails
                details={{
                    positionValue: 34,
                    nominalUnlockedValue: 17,
                    executableSaleValue: 0,
                }}
            />
        );

        const detailsPanel = screen.getByTestId('projection-error-details');
        expect(detailsPanel.className).toContain('space-y-2');
        expect(detailsPanel.className).not.toContain('sm:grid-cols-3');

        const executableRow = screen.getByTestId('projection-error-detail-currently-executable-sale-value');
        expect(within(executableRow).getByText('Currently Executable Sale Value').className).toContain('break-words');
        expect(within(executableRow).getByText('0')).not.toBeNull();
    });
});
