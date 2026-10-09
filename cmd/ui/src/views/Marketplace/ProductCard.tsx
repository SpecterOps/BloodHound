// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { Badge, ButtonVariants, Card, CardContent, Link, Typography } from 'doodle-ui';

export type ProductCardProps = {
    name: string;
    description: string;
    href: string;
    linkLabel: string;
    logoPath: string;
    author: string;
    categoryLabel: string;
    badge?: string;
};

const ProductCard: React.FC<ProductCardProps> = ({
    name,
    description,
    href,
    linkLabel,
    logoPath,
    author,
    categoryLabel,
    badge,
}) => {
    return (
        <Card
            aria-label={name}
            className='flex h-full flex-col border border-neutral-light-4 bg-white shadow-none transition-shadow hover:shadow-outer-1 dark:border-neutral-dark-4 dark:bg-neutral-dark-2'
            role='article'>
            <CardContent className='flex flex-1 flex-col gap-4 p-5'>
                <div className='flex items-start gap-4'>
                    <div className='flex size-16 shrink-0 items-center justify-center rounded-lg border border-neutral-light-4 bg-white p-2 shadow-outer-1'>
                        <img
                            alt=''
                            className='max-h-full max-w-full object-contain'
                            data-testid='product-logo'
                            src={`${import.meta.env.BASE_URL.replace(/\/$/, '')}${logoPath}`}
                        />
                    </div>
                    <div className='min-w-0 flex-1'>
                        <div className='flex items-start justify-between gap-2'>
                            <Typography component='h3' variant='h3'>
                                {name}
                            </Typography>
                            {badge && <Badge className='shrink-0' color='blue' label={badge} variant='fill' />}
                        </div>
                        <Typography className='mt-1 text-muted-foreground opacity-75' variant='body2'>
                            {author}
                        </Typography>
                    </div>
                </div>
                <Typography className='flex-1 text-muted-foreground' variant='body1'>
                    {description}
                </Typography>
                <div className='flex flex-wrap items-center justify-between gap-3'>
                    <Typography className='text-muted-foreground opacity-75' variant='body2'>
                        {categoryLabel}
                    </Typography>
                    <Link
                        aria-label={`${linkLabel} ${name} (opens in a new tab)`}
                        className={`${ButtonVariants({ variant: 'secondary' })} rounded-md font-bold`}
                        href={href}
                        variant='unstyled'>
                        {linkLabel}
                    </Link>
                </div>
            </CardContent>
        </Card>
    );
};

export default ProductCard;
