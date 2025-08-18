import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { render, screen } from '../../../test-utils';
import TagToZoneLabel from './TagToZoneLabel';

const testSelectedQuery = {
    name: '10 Admins',
    description: '10 Admins',
    query: "MATCH p = (t:Group)<-[:MemberOf*1..]-(a)\nWHERE (a:User or a:Computer) and t.objectid ENDS WITH '-512'\nRETURN p\nLIMIT 10",
    canEdit: true,
    id: 1,
    user_id: '4e09c965-65bd-4f15-ae71-5075a6fed14b',
};

const handlers = [
    rest.get('/api/v2/asset-group-tags', async (_, res, ctx) => {
        return res(
            ctx.json({
                data: {},
            })
        );
    }),
];

const server = setupServer(...handlers);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('TagToZoneLabel', () => {
    it('renders trigger button and dropdown items', async () => {
        const user = userEvent.setup();

        render(<TagToZoneLabel selectedQuery={testSelectedQuery} cypherQuery={''}></TagToZoneLabel>);

        const tagTrigger = screen.getByText('Tag');
        expect(tagTrigger).toBeInTheDocument();

        await user.click(tagTrigger);

        expect(screen.getByText('Zone')).toBeInTheDocument();
        expect(screen.getByText('Label')).toBeInTheDocument();
    });

    it('does fires TagToZoneDialog when the Zone option is clicked', async () => {
        const user = userEvent.setup();

        render(<TagToZoneLabel selectedQuery={testSelectedQuery} cypherQuery={''}></TagToZoneLabel>);

        const tagTrigger = screen.getByText('Tag');
        expect(tagTrigger).toBeInTheDocument();

        await user.click(tagTrigger);
        const zoneOption = screen.getByText('Zone');

        await user.click(zoneOption);

        expect(zoneOption).not.toBeInTheDocument();
        expect(screen.getByText('Tag Results to Zone')).toBeInTheDocument();
    });

    it('does fires TagToLabelDialog when the Label option is clicked', async () => {
        const user = userEvent.setup();

        render(<TagToZoneLabel selectedQuery={testSelectedQuery} cypherQuery={''}></TagToZoneLabel>);

        const tagTrigger = screen.getByText('Tag');
        expect(tagTrigger).toBeInTheDocument();

        await user.click(tagTrigger);
        const labelOption = screen.getByText('Label');

        await user.click(labelOption);

        expect(labelOption).not.toBeInTheDocument();
        expect(screen.getByText('Tag Results to Label')).toBeInTheDocument();
    });
});
