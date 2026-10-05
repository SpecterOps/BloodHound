import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{B as c}from"./Button-DTtPuz4n.js";import{C as M,a as E,b as _,c as F}from"./Card-BdBAz3fL.js";import{I as U}from"./Input--sj12WP2.js";import{L as q}from"./Link-Q3xt4pq6.js";import{t as I,T as a,v as p}from"./Typography-BMbK6Q6M.js";import"./index-ChsGqxH_.js";import"./index-Bb4qSo10.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./Label-Dsc1dydc.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-TwK_z8r1.js";import"./index.es-C0EmWI2V.js";const ie={title:"Components/Typography",component:a,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{variant:{type:"string",options:Object.keys(p),control:"select",description:"Applies default styling based on heading/tag level:"},component:{options:I,control:"select",description:"Applies selected html tag. Overrides default tag from variant mapping."}},args:{}},t={name:"Typography",args:{variant:"h1",children:"The quick brown fox jumps over the lazy dog"},render:r=>e.jsx(a,{...r})},n={render:()=>e.jsx("div",{className:"w-[40rem] max-w-full space-y-6",children:Object.keys(p).map(r=>e.jsxs("div",{className:"grid grid-cols-[10rem_1fr] items-baseline gap-6",children:[e.jsxs(a,{variant:"caption",className:"text-text-muted",children:[r," · <",p[r],">"]}),e.jsx(a,{variant:r,children:"The quick brown fox jumps over the lazy dog"})]},r))})},s={render:()=>e.jsxs("article",{className:"w-96 max-w-full space-y-3",children:[e.jsx(a,{variant:"h2",children:"Review attack path exposure"}),e.jsx(a,{variant:"subtitle1",children:"A summary of the highest-risk paths in your environment"}),e.jsx(a,{variant:"body1",children:"Attack paths chain together the misconfigurations and privileges that let a principal move toward your most critical assets. Prioritize the paths that expose Tier Zero, then work outward."}),e.jsx(a,{variant:"body2",children:"Longer explanatory copy stays comfortable to read across several lines, preserving spacing and line height even when the available width is reduced by browser zoom or a narrow container."}),e.jsx(a,{variant:"caption",children:"Last calculated 2 minutes ago"})]})},o={render:()=>e.jsx("div",{className:"w-[36rem] max-w-full divide-y divide-neutral-400 rounded border border-neutral-400",children:["Domain Admins","Enterprise Admins","Remote Desktop Users","Backup Operators"].map((r,d)=>e.jsxs("div",{className:"grid grid-cols-[1fr_auto] gap-4 p-3",children:[e.jsxs("div",{children:[e.jsx(a,{variant:"subtitle2",children:r}),e.jsxs(a,{variant:"body2",children:["Active Directory group · Tier ",d%2]})]}),e.jsxs(a,{variant:"caption",children:[12+d*7," members"]})]},r))})},i={render:()=>e.jsxs(M,{className:"w-[28rem] max-w-full",children:[e.jsxs(E,{children:[e.jsx(a,{variant:"h3",children:"Review attack-path exposure"}),e.jsx(a,{variant:"subtitle2",children:"Updated a few seconds ago"})]}),e.jsxs(_,{className:"space-y-3",children:[e.jsx(a,{variant:"body1",children:"Prioritize the paths that give principals control of your most critical assets."}),e.jsxs(a,{variant:"body2",children:["Learn how exposure is calculated in the"," ",e.jsx(q,{href:"https://bloodhound.specterops.io/",className:"inline-flex",children:"BloodHound documentation"}),"."]}),e.jsxs("div",{children:[e.jsx(a,{id:"typography-card-filter-label",variant:"caption",component:"span",children:"Filter by asset name"}),e.jsx(U,{id:"typography-card-filter","aria-labelledby":"typography-card-filter-label",className:"mt-1",placeholder:"Search assets"})]})]}),e.jsxs(F,{className:"gap-2",children:[e.jsx(c,{size:"small",children:"Review paths"}),e.jsx(c,{size:"small",variant:"secondary",children:"Dismiss"})]})]})};var l,h,m,y,g;t.parameters={...t.parameters,docs:{...(l=t.parameters)==null?void 0:l.docs,source:{originalSource:`{
  name: 'Typography',
  args: {
    variant: 'h1',
    children: 'The quick brown fox jumps over the lazy dog'
  },
  render: args => <Typography {...args} />
}`,...(m=(h=t.parameters)==null?void 0:h.docs)==null?void 0:m.source},description:{story:"The default story. Use the controls to change the `variant` and, optionally, the rendered\n`component`. Each `variant` applies a preset type style and a matching default HTML element; the\n`component` prop overrides that element without changing the visual style.\n\n```tsx\n<Typography variant='h1'>Attack path exposure</Typography>\n<Typography variant='h1' component='span'>Same style, rendered as a span</Typography>\n```",...(g=(y=t.parameters)==null?void 0:y.docs)==null?void 0:g.description}}};var u,v,x,b,T;n.parameters={...n.parameters,docs:{...(u=n.parameters)==null?void 0:u.docs,source:{originalSource:`{
  render: () => <div className='w-[40rem] max-w-full space-y-6'>
            {(Object.keys(variantMapping) as Variant[]).map(variant => <div className='grid grid-cols-[10rem_1fr] items-baseline gap-6' key={variant}>
                    <Typography variant='caption' className='text-text-muted'>
                        {variant} · &lt;{variantMapping[variant]}&gt;
                    </Typography>
                    <Typography variant={variant}>The quick brown fox jumps over the lazy dog</Typography>
                </div>)}
        </div>
}`,...(x=(v=n.parameters)==null?void 0:v.docs)==null?void 0:x.source},description:{story:"The complete type scale. Each row pairs the `variant` token and the HTML element it renders by\ndefault with a specimen of the style, making it easy to compare size, weight, and spacing.",...(T=(b=n.parameters)==null?void 0:b.docs)==null?void 0:T.description}}};var w,f,j,k,C;s.parameters={...s.parameters,docs:{...(w=s.parameters)==null?void 0:w.docs,source:{originalSource:`{
  render: () => <article className='w-96 max-w-full space-y-3'>
            <Typography variant='h2'>Review attack path exposure</Typography>
            <Typography variant='subtitle1'>A summary of the highest-risk paths in your environment</Typography>
            <Typography variant='body1'>
                Attack paths chain together the misconfigurations and privileges that let a principal move toward your
                most critical assets. Prioritize the paths that expose Tier Zero, then work outward.
            </Typography>
            <Typography variant='body2'>
                Longer explanatory copy stays comfortable to read across several lines, preserving spacing and line
                height even when the available width is reduced by browser zoom or a narrow container.
            </Typography>
            <Typography variant='caption'>Last calculated 2 minutes ago</Typography>
        </article>
}`,...(j=(f=s.parameters)==null?void 0:f.docs)==null?void 0:j.source},description:{story:`How the variants read together in flowing, real-world content. Text wraps naturally and preserves
its size and line height across multiple lines and constrained widths.`,...(C=(k=s.parameters)==null?void 0:k.docs)==null?void 0:C.description}}};var N,L,z,A,B;o.parameters={...o.parameters,docs:{...(N=o.parameters)==null?void 0:N.docs,source:{originalSource:`{
  render: () => <div className='w-[36rem] max-w-full divide-y divide-neutral-400 rounded border border-neutral-400'>
            {['Domain Admins', 'Enterprise Admins', 'Remote Desktop Users', 'Backup Operators'].map((name, index) => <div className='grid grid-cols-[1fr_auto] gap-4 p-3' key={name}>
                    <div>
                        <Typography variant='subtitle2'>{name}</Typography>
                        <Typography variant='body2'>Active Directory group · Tier {index % 2}</Typography>
                    </div>
                    <Typography variant='caption'>{12 + index * 7} members</Typography>
                </div>)}
        </div>
}`,...(z=(L=o.parameters)==null?void 0:L.docs)==null?void 0:z.source},description:{story:"A compact, information-dense layout that combines `subtitle2`, `body2`, and `caption` to present\nmany rows of related data, such as a list of groups or assets.",...(B=(A=o.parameters)==null?void 0:A.docs)==null?void 0:B.description}}};var D,R,H,O,S;i.parameters={...i.parameters,docs:{...(D=i.parameters)==null?void 0:D.docs,source:{originalSource:`{
  render: () => <Card className='w-[28rem] max-w-full'>
            <CardHeader>
                <Typography variant='h3'>Review attack-path exposure</Typography>
                <Typography variant='subtitle2'>Updated a few seconds ago</Typography>
            </CardHeader>
            <CardContent className='space-y-3'>
                <Typography variant='body1'>
                    Prioritize the paths that give principals control of your most critical assets.
                </Typography>
                <Typography variant='body2'>
                    Learn how exposure is calculated in the{' '}
                    <Link href='https://bloodhound.specterops.io/' className='inline-flex'>
                        BloodHound documentation
                    </Link>
                    .
                </Typography>
                <div>
                    <Typography id='typography-card-filter-label' variant='caption' component='span'>
                        Filter by asset name
                    </Typography>
                    <Input id='typography-card-filter' aria-labelledby='typography-card-filter-label' className='mt-1' placeholder='Search assets' />
                </div>
            </CardContent>
            <CardFooter className='gap-2'>
                <Button size='small'>Review paths</Button>
                <Button size='small' variant='secondary'>
                    Dismiss
                </Button>
            </CardFooter>
        </Card>
}`,...(H=(R=i.parameters)==null?void 0:R.docs)==null?void 0:H.source},description:{story:`Typography composed with other components inside a Card: a heading and timestamp, supporting body
copy with an inline Link, and a caption used as an accessible label for an Input.`,...(S=(O=i.parameters)==null?void 0:O.docs)==null?void 0:S.description}}};const pe=["TypographyComponent","Variants","MultilineAndLongText","DenseList","CardComposition"];export{i as CardComposition,o as DenseList,s as MultilineAndLongText,t as TypographyComponent,n as Variants,pe as __namedExportsOrder,ie as default};
