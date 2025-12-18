const ShortcutSection = ({ heading, bindings }: { heading: string; bindings: string[][] }) => (
    <div className='mb-5' key={heading}>
        <div className='font-bold flex sm:justify-center p-2'>{heading}</div>
        <div className='flex flex-col gap-2 text-sm'>
            {bindings.map((binding: string[]) => (
                <div key={`${binding[0]}-${heading}`} className='flex gap-2'>
                    <div className='w-1/2 text-right p-2 flex md:justify-end sm:justify-center xs:justify-center items-center'>
                        {' '}
                        {binding[1]}
                    </div>
                    <div className='w-1/2 border-2 rounded-md p-2 text-center flex justify-center items-center'>
                        Alt/Option + {binding[0]}
                    </div>
                </div>
            ))}
        </div>
    </div>
);

export default ShortcutSection;
