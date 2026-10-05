import{j as t}from"./jsx-runtime-BTJTZTIL.js";import{T as m,a as p,b as T,c as r,d as l,e as b,f as s,g as u,r as e}from"./index-tgJspgZb.js";import"./index-ChsGqxH_.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index.es-C0EmWI2V.js";import"./IconButton-CiUmUVXd.js";import"./index-Bb4qSo10.js";import"./AccessibleIcon-BlGV1aYS.js";import"./Tooltip-4xw-YYUR.js";import"./index-Dv0R_RYs.js";import"./index-Dc_EP5I0.js";import"./index-BWEBPL6D.js";import"./index-DnVAETpg.js";import"./index-ejpvWIAw.js";import"./index-DSMk1t8m.js";import"./floating-ui.react-dom-D40P0bPq.js";import"./index-DRLFypRQ.js";import"./index-vQUXxau3.js";import"./index-oIeaYTeP.js";import"./AppIcons-Dp2sojKC.js";import"./Card-BdBAz3fL.js";import"./Typography-BMbK6Q6M.js";import"./utils-BIldDEfN.js";import"./UserCog-VADywgBc.js";import"./VisuallyHidden-DT_8ZbEK.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";const Q={title:"Components/Table",component:m,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{},args:{}},i=[{id:e.string.uuid(),todo:e.hacker.phrase(),status:"complete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"incomplete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"complete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"complete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"incomplete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"incomplete",description:e.lorem.sentence()},{id:e.string.uuid(),todo:e.hacker.phrase(),status:"complete",description:e.lorem.sentence()}],o={render:()=>t.jsxs(m,{className:"w-full",children:[t.jsx(p,{children:"A list of your recent TODOs."}),t.jsx(T,{children:t.jsxs(r,{children:[t.jsx(l,{className:"w-48",children:"Todo"}),t.jsx(l,{className:"w-48",children:"Description"}),t.jsx(l,{className:"w-32",children:"Status"})]})}),t.jsx(b,{children:i.map(a=>t.jsxs(r,{children:[t.jsx(s,{className:"font-medium",children:a.todo}),t.jsx(s,{children:a.description}),t.jsx(s,{children:a.status})]},a.id))}),t.jsx(u,{children:t.jsxs(r,{children:[t.jsx(s,{colSpan:3,children:"Completed"}),t.jsx(s,{className:"text-right",children:i.filter(a=>a.status==="complete").length})]})})]})};var n,c,d;o.parameters={...o.parameters,docs:{...(n=o.parameters)==null?void 0:n.docs,source:{originalSource:`{
  render: () => <Table className='w-full'>
            <TableCaption>A list of your recent TODOs.</TableCaption>
            <TableHeader>
                <TableRow>
                    <TableHead className='w-48'>Todo</TableHead>
                    <TableHead className='w-48'>Description</TableHead>
                    <TableHead className='w-32'>Status</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {invoices.map(invoice => <TableRow key={invoice.id}>
                        <TableCell className='font-medium'>{invoice.todo}</TableCell>
                        <TableCell>{invoice.description}</TableCell>
                        <TableCell>{invoice.status}</TableCell>
                    </TableRow>)}
            </TableBody>
            <TableFooter>
                <TableRow>
                    <TableCell colSpan={3}>Completed</TableCell>
                    <TableCell className='text-right'>
                        {invoices.filter(s => s.status === 'complete').length}
                    </TableCell>
                </TableRow>
            </TableFooter>
        </Table>
}`,...(d=(c=o.parameters)==null?void 0:c.docs)==null?void 0:d.source}}};const U=["Default"];export{o as Default,U as __namedExportsOrder,Q as default};
