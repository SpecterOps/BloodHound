import { createColumnHelper, DataTable } from '@bloodhoundenterprise/doodleui';
import {
    CertificationAuto,
    CertificationManual,
    CertificationPending,
    CertificationRevoked,
    CertificationTypeMap,
} from 'js-client-library';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect, useMemo, useRef } from 'react';
import { AppIcon, DropdownOption, DropdownSelector } from '../../../components';
import { useAssetGroupTags, useAvailableEnvironments } from '../../../hooks';

type CertificationTableProps = {
    data: any;
    isLoading: boolean;
    isFetching: boolean;
    isSuccess: boolean;
    fetchNextPage: any;
    selectedRows: number[];
    setSelectedRows: any;
};

const CertificationTable: FC<CertificationTableProps> = ({
    data,
    isLoading,
    isFetching,
    isSuccess,
    fetchNextPage,
    selectedRows,
    setSelectedRows,
}) => {
    const mockPending = '9';
    const scrollRef = useRef<HTMLDivElement>(null);

    const { data: availableEnvironments = [] } = useAvailableEnvironments();
    const { data: assetGroupTags = [] } = useAssetGroupTags();

    const domainMap = useMemo(() => {
        const map = new Map<string, string>();
        for (const domain of availableEnvironments) {
            map.set(domain.id, domain.name);
        }
        return map;
    }, [availableEnvironments]);

    const tagMap = useMemo(() => {
        const map = new Map<number, string>();
        for (const tag of assetGroupTags) {
            map.set(tag.id, tag.name);
        }
        return map;
    }, [assetGroupTags]);

    const certificationsData = data ?? { pages: [{ count: 0, data: { members: [] } }] };
    const totalDBRowCount = certificationsData.pages[0].count;
    const certificationsItemsRaw = certificationsData.pages.flatMap((item) => item.data.members);
    const totalFetched = certificationsItemsRaw.length;

    const fetchMoreOnBottomReached = useCallback(
        (containerRefElement?: HTMLDivElement | null) => {
            if (containerRefElement) {
                const { scrollHeight, scrollTop, clientHeight } = containerRefElement;
                //once the user has scrolled within 500px of the bottom of the table, fetch more data if we can
                if (scrollHeight - scrollTop - clientHeight < 500 && !isFetching && totalFetched < totalDBRowCount) {
                    fetchNextPage();
                }
            }
        },
        [fetchNextPage, isFetching, totalFetched, totalDBRowCount]
    );

    useEffect(() => {
        fetchMoreOnBottomReached(scrollRef.current);
    }, [fetchMoreOnBottomReached]);

    const toggleRow = (id: number) => {
        setSelectedRows((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
    };

    const toggleAll = () => {
        if (selectedRows.length === data.length) {
            setSelectedRows([]);
        } else {
            setSelectedRows(data.map((row) => row.id));
        }
    };

    const allSelected = selectedRows.length === data?.length;
    const someSelected = selectedRows.length > 0 && !allSelected;

    const columnHelper = createColumnHelper<any>();

    const columns = [
        columnHelper.display({
            id: 'bulk-certify',
            header: () => (
                <div className='pl-8'>
                    <input
                        type='checkbox'
                        checked={allSelected}
                        ref={(el) => {
                            if (el) el.indeterminate = someSelected;
                        }}
                        onChange={toggleAll}
                    />
                </div>
            ),
            cell: (info) => (
                <div className='pl-8'>
                    <input
                        type='checkbox'
                        checked={selectedRows.includes(info.row.original.id)}
                        onChange={() => toggleRow(info.row.original.id)}
                    />
                </div>
            ),
        }),
        columnHelper.accessor('primary_kind', {
            header: () => <div className='pl-8 text-left'>Type</div>,
            cell: (info) => <div className='text-primary dark:text-secondary-variant-2'>{info.getValue()}</div>,
        }),
        columnHelper.accessor('name', {
            header: () => <div className='pl-8 text-left'>Member Name</div>,
            cell: (info) => <div>{info.getValue()}</div>,
        }),
        columnHelper.accessor('domainName', {
            header: () => <div className='pl-8 text-left'>Domain</div>,
            cell: (info) => <div>{info.getValue()}</div>,
        }),
        columnHelper.accessor('zoneName', {
            header: () => <div className='pl-8 text-left'>Zone</div>,
            cell: (info) => <div>{info.getValue()}</div>,
        }),
        columnHelper.accessor('date', {
            header: () => <div className='pl-8 text-center'>First Seen</div>,
            cell: (info) => <div className='text-center'>{info.getValue()}</div>,
        }),
    ];

    type DataTableProps = React.ComponentProps<typeof DataTable>;

    const tableProps: DataTableProps['TableProps'] = {
        className: 'table-fixed',
        disableDefaultOverflowAuto: true,
    };

    const tableHeaderProps: DataTableProps['TableHeaderProps'] = {
        className: 'sticky top-0 z-10 shadow-sm text-base',
    };

    const tableHeadProps: DataTableProps['TableHeadProps'] = {
        className: 'pr-2 text-center',
    };

    const tableCellProps: DataTableProps['TableCellProps'] = {
        className: 'truncate group relative pl-8',
    };

    const virtualizationOptions: DataTableProps['virtualizationOptions'] = {
        estimateSize: () => 79,
    };

    const certOptions: DropdownOption[] = [
        CertificationPending,
        CertificationManual,
        CertificationAuto,
        CertificationRevoked,
    ].map((certType) => {
        return { key: certType, value: CertificationTypeMap[certType] };
    });

    const certificationsItems = isSuccess
        ? certificationsItemsRaw.map((item) => {
              return {
                  ...item,
                  date: DateTime.fromISO(item.created_at).toFormat('MM-dd-yyyy'),
                  domainName: domainMap.get(item.environment_id) ?? 'Unknown',
                  zoneName: tagMap.get(item.asset_group_tag_id) ?? 'Unknown',
              };
          })
        : [];

    console.log('selected rows in child!!', selectedRows);

    return (
        <div className='bg-neutral-light-2 dark:bg-neutral-dark-2'>
            <div className='flex items-center'>
                <h1 className='text-xl font-bold'>Certifications</h1>
                <p>{`${mockPending} pending`}</p>
            </div>
            <div>
                <DropdownSelector
                    variant='transparent'
                    options={certOptions}
                    selectedText={
                        <span className='flex items-center gap-3'>
                            <AppIcon.CertStatus size={24} /> Status
                        </span>
                    }
                    onChange={(_selectedCertificationType: DropdownOption) => {}}></DropdownSelector>
            </div>
            <DataTable
                data={certificationsItems ?? []}
                TableHeaderProps={tableHeaderProps}
                TableHeadProps={tableHeadProps}
                TableProps={tableProps}
                TableCellProps={tableCellProps}
                columns={columns}
                virtualizationOptions={virtualizationOptions}
            />
        </div>
    );
};

export default CertificationTable;
