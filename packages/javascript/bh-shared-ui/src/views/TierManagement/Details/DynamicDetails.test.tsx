import { AssetGroupTagTypeValues } from 'js-client-library';
import { render, screen } from '../../../test-utils';
import DynamicDetails from './DynamicDetails';

describe('DynamicDetails', () => {
    it('renders details for a selected tier', () => {
        const testQuery = {
            asset_group_tier_id: 9,
            count: 9374,
            requireCertify: true,
            created_at: '2024-09-08T03:38:22.791Z',
            created_by: 'Franz.Smitham@yahoo.com',
            deleted_at: '2025-02-03T18:32:36.669Z',
            deleted_by: 'Vita.Hermann97@yahoo.com',
            description: 'pique International',
            id: 9,
            kind_id: 59514,
            name: 'Tier-8',
            updated_at: '2024-07-26T02:15:04.556Z',
            updated_by: 'Deontae34@hotmail.com',
            position: 0,
            type: 1 as AssetGroupTagTypeValues,
        };

        render(<DynamicDetails data={testQuery} />);

        expect(screen.getByText('Tier-8')).toBeInTheDocument();
        expect(screen.getByText('pique International')).toBeInTheDocument();
        expect(screen.getByText('Franz.Smitham@yahoo.com')).toBeInTheDocument();
        expect(screen.getByText('7/25/2024')).toBeInTheDocument();
    });

    it('renders details for a selected selector and is of type "Cypher"', () => {
        const testQuery = {
            asset_group_tag_id: 9,
            allow_disable: false,
            selector_id: 1,
            node_id: 1,
            certified: 1,
            certified_by: 'Test',
            auto_certify: true,
            count: 67369,
            created_at: '2025-02-12T16:24:18.633Z',
            created_by: 'Emery_Swift86@gmail.com',
            description: 'North',
            disabled_at: '2024-05-24T12:34:35.894Z',
            disabled_by: 'Travon27@gmail.com',
            id: 9,
            is_default: true,
            seeds: [],
            name: 'tier-0-selector-9',
            updated_at: '2024-11-25T11:34:45.894Z',
            updated_by: 'Demario_Corwin88@yahoo.com',
        };

        render(<DynamicDetails data={testQuery} isCypher={true} />);

        expect(screen.getByText('tier-0-selector-9')).toBeInTheDocument();
        expect(screen.getByText('North')).toBeInTheDocument();
        expect(screen.getByText('Emery_Swift86@gmail.com')).toBeInTheDocument();
        expect(screen.getByText('11/25/2024')).toBeInTheDocument();
        expect(screen.getByText('Cypher')).toBeInTheDocument();
    });

    it('renders details for a selected selector and is of type "Object"', () => {
        const testQuery = {
            asset_group_tag_id: 9,
            allow_disable: false,
            selector_id: 1,
            node_id: 1,
            certified: 1,
            certified_by: 'Test',
            auto_certify: true,
            count: 67369,
            seeds: [],
            created_at: '2025-02-12T16:24:18.633Z',
            created_by: 'Emery_Swift86@gmail.com',
            description: 'North',
            disabled_at: '2024-05-24T12:34:35.894Z',
            disabled_by: 'Travon27@gmail.com',
            id: 9,
            is_default: true,
            name: 'tier-0-selector-9',
            updated_at: '2024-11-25T11:34:45.894Z',
            updated_by: 'Demario_Corwin88@yahoo.com',
        };

        render(<DynamicDetails data={testQuery} />);

        expect(screen.getByText('tier-0-selector-9')).toBeInTheDocument();
        expect(screen.getByText('North')).toBeInTheDocument();
        expect(screen.getByText('Emery_Swift86@gmail.com')).toBeInTheDocument();
        expect(screen.getByText('11/25/2024')).toBeInTheDocument();
        expect(screen.getByText('Object')).toBeInTheDocument();
    });
});
