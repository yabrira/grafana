import { render, screen } from '@testing-library/react';

import { DebugDumpPreview } from './DebugDumpPreview';

describe('DebugDumpPreview', () => {
  it('renders the dump title and html payload', () => {
    render(<DebugDumpPreview title="Bundle notes" html="<b>collected</b>" />);

    expect(screen.getByRole('heading', { name: 'Bundle notes' })).toBeInTheDocument();
    expect(screen.getByText('collected')).toBeInTheDocument();
  });
});
