import{j as a}from"./jsx-runtime-BTJTZTIL.js";import{C as r,a as c}from"./Checkbox-Ct8FIfJk.js";import"./index-ChsGqxH_.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-Dv0R_RYs.js";import"./index-CFknkR0k.js";import"./index-DRLFypRQ.js";import"./index-vQUXxau3.js";import"./index-Bb4qSo10.js";import"./Label-Dsc1dydc.js";import"./index-TwK_z8r1.js";import"./utils-BIldDEfN.js";import"./Typography-BMbK6Q6M.js";import"./createLucideIcon-rMs-1iam.js";const ee={title:"Components/Checkbox",component:r,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{size:{control:"select",options:["lg","md","sm"],description:"Size of the Checkbox."},disabled:{control:"boolean",description:"Disables interaction with the Checkbox."}},args:{size:"md",disabled:!1}},t={render:e=>a.jsx(r,{"aria-label":"Checkbox",...e})},l={render:e=>a.jsx(c,{"aria-label":"Checkbox with label",...e,label:"Label",defaultChecked:!1})},i={render:({size:e})=>a.jsxs("div",{className:"flex flex-col items-start gap-4",children:[a.jsx(c,{size:e,label:"Unchecked",checked:!1}),a.jsx(c,{size:e,label:"Checked",checked:!0}),a.jsx(c,{size:e,label:"Indeterminate",checked:"indeterminate"}),a.jsx(c,{size:e,label:"Error",error:!0,checked:!0}),a.jsx(c,{size:e,label:"Disabled",disabled:!0,checked:!0})]})},s={render:e=>a.jsxs("div",{className:"flex items-center gap-2",children:[a.jsx(r,{"aria-label":"Unchecked checkbox",checked:!1,...e}),a.jsx(r,{"aria-label":"Checked checkbox",checked:!0,...e}),a.jsx(r,{"aria-label":"Indeterminate checkbox",checked:"indeterminate",...e})]})},d={render:e=>a.jsxs("div",{className:"flex items-center gap-2",children:[a.jsx(r,{"aria-label":"Disabled unchecked checkbox",id:"checkbox-disabled-unchecked",checked:!1,...e,disabled:!0}),a.jsx(r,{"aria-label":"Disabled checked checkbox",id:"checkbox-disabled-checked",checked:!0,...e,disabled:!0}),a.jsx(r,{"aria-label":"Disabled indeterminate checkbox",id:"checkbox-disabled-indeterminate",checked:"indeterminate",...e,disabled:!0})]})},o={render:e=>a.jsxs("div",{className:"flex items-center gap-2",children:[a.jsx(r,{"aria-label":"Error unchecked checkbox",id:"checkbox-error-unchecked",checked:!1,...e,"aria-invalid":!0}),a.jsx(r,{"aria-label":"Error checked checkbox",id:"checkbox-error-checked",checked:!0,...e,"aria-invalid":!0}),a.jsx(r,{"aria-label":"Error indeterminate checkbox",id:"checkbox-error-indeterminate",checked:"indeterminate",...e,"aria-invalid":!0})]})};var n,b,h;t.parameters={...t.parameters,docs:{...(n=t.parameters)==null?void 0:n.docs,source:{originalSource:`{
  render: args => <Checkbox aria-label='Checkbox' {...args} />
}`,...(h=(b=t.parameters)==null?void 0:b.docs)==null?void 0:h.source}}};var k,x,m;l.parameters={...l.parameters,docs:{...(k=l.parameters)==null?void 0:k.docs,source:{originalSource:`{
  render: args => <CheckboxWithLabel aria-label='Checkbox with label' {...args} label='Label' defaultChecked={false} />
}`,...(m=(x=l.parameters)==null?void 0:x.docs)==null?void 0:m.source}}};var p,u,C,f,g;i.parameters={...i.parameters,docs:{...(p=i.parameters)==null?void 0:p.docs,source:{originalSource:`{
  render: ({
    size
  }) => <div className='flex flex-col items-start gap-4'>
            <CheckboxWithLabel size={size} label='Unchecked' checked={false} />
            <CheckboxWithLabel size={size} label='Checked' checked />
            <CheckboxWithLabel size={size} label='Indeterminate' checked='indeterminate' />
            <CheckboxWithLabel size={size} label='Error' error checked />
            <CheckboxWithLabel size={size} label='Disabled' disabled checked />
        </div>
}`,...(C=(u=i.parameters)==null?void 0:u.docs)==null?void 0:C.source},description:{story:"Checkboxes with a label.",...(g=(f=i.parameters)==null?void 0:f.docs)==null?void 0:g.description}}};var j,v,E,L,D;s.parameters={...s.parameters,docs:{...(j=s.parameters)==null?void 0:j.docs,source:{originalSource:`{
  render: args => <div className='flex items-center gap-2'>
            <Checkbox aria-label='Unchecked checkbox' checked={false} {...args} />
            <Checkbox aria-label='Checked checkbox' checked={true} {...args} />
            <Checkbox aria-label='Indeterminate checkbox' checked='indeterminate' {...args} />
        </div>
}`,...(E=(v=s.parameters)==null?void 0:v.docs)==null?void 0:E.source},description:{story:"Icons only Checkboxes.",...(D=(L=s.parameters)==null?void 0:L.docs)==null?void 0:D.description}}};var W,z,N,y,S;d.parameters={...d.parameters,docs:{...(W=d.parameters)==null?void 0:W.docs,source:{originalSource:`{
  render: args => <div className='flex items-center gap-2'>
            <Checkbox aria-label='Disabled unchecked checkbox' id='checkbox-disabled-unchecked' checked={false} {...args} disabled />
            <Checkbox aria-label='Disabled checked checkbox' id='checkbox-disabled-checked' checked {...args} disabled />
            <Checkbox aria-label='Disabled indeterminate checkbox' id='checkbox-disabled-indeterminate' checked='indeterminate' {...args} disabled />
        </div>
}`,...(N=(z=d.parameters)==null?void 0:z.docs)==null?void 0:N.source},description:{story:"Disabled Checkbox cannot be interacted with.",...(S=(y=d.parameters)==null?void 0:y.docs)==null?void 0:S.description}}};var w,I,U,_,O;o.parameters={...o.parameters,docs:{...(w=o.parameters)==null?void 0:w.docs,source:{originalSource:`{
  render: args => <div className='flex items-center gap-2'>
            <Checkbox aria-label='Error unchecked checkbox' id='checkbox-error-unchecked' checked={false} {...args} aria-invalid />
            <Checkbox aria-label='Error checked checkbox' id='checkbox-error-checked' checked={true} {...args} aria-invalid />
            <Checkbox aria-label='Error indeterminate checkbox' id='checkbox-error-indeterminate' checked='indeterminate' {...args} aria-invalid />
        </div>
}`,...(U=(I=o.parameters)==null?void 0:I.docs)==null?void 0:U.source},description:{story:"Error Checkboxes.",...(O=(_=o.parameters)==null?void 0:_.docs)==null?void 0:O.description}}};const ae=["Default","WithLabel","LabelExamples","CheckboxWithoutLabel","Disabled","Error"];export{s as CheckboxWithoutLabel,t as Default,d as Disabled,o as Error,i as LabelExamples,l as WithLabel,ae as __namedExportsOrder,ee as default};
