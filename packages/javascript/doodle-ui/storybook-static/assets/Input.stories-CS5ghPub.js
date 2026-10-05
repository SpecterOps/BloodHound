import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{B as Z}from"./Button-DTtPuz4n.js";import{L as Y}from"./Label-Dsc1dydc.js";import{I as p}from"./Input--sj12WP2.js";import"./index-ChsGqxH_.js";import"./index-Bb4qSo10.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-TwK_z8r1.js";import"./Typography-BMbK6Q6M.js";const ge={title:"Components/Input",component:p,tags:["autodocs"],argTypes:{variant:{options:["outlined","underlined"],control:"select"}},args:{className:"w-96",type:"email",placeholder:"email@example.com",disabled:!1,variant:"outlined"},parameters:{layout:"centered",docs:{description:{component:"Displays a form input field or a component that looks like an input field."}}}},a={},o={args:{label:"Email address",required:!0,variant:"outlined"}},n={args:{required:!0}},d={args:{defaultValue:"not-an-email",error:!0,errorMessage:"Enter a valid email address.",label:"Email address",required:!0,variant:"outlined"}},t={args:{disabled:!0,variant:"outlined"}},l={args:{id:"email",label:"Email"}},s={args:{id:"email-2",label:"Email",helperText:"Enter your email address.",required:!0,variant:"outlined"}},i={render:r=>e.jsxs("div",{className:"flex items-center space-x-2",children:[e.jsx(p,{...r}),e.jsx(Z,{className:"rounded-md",children:"Subscribe"})]})},c={args:{type:"file",placeholder:"No file selected."},render:r=>e.jsxs("div",{className:"grid w-full max-w-sm items-center gap-1.5",children:[e.jsx(Y,{htmlFor:"picture",children:"Upload Image"}),e.jsx(p,{id:"picture",type:"file",...r,variant:"outlined"})]})},m={args:{intent:"time",type:"time"},render:r=>e.jsxs("div",{className:"grid items-center gap-1.5",children:[e.jsx(Y,{htmlFor:"time",children:"Time"}),e.jsx(p,{...r,id:"time"})]})};var u,g,h,f,b;a.parameters={...a.parameters,docs:{...(u=a.parameters)==null?void 0:u.docs,source:{originalSource:"{}",...(h=(g=a.parameters)==null?void 0:g.docs)==null?void 0:h.source},description:{story:"The default form of the input field.",...(b=(f=a.parameters)==null?void 0:f.docs)==null?void 0:b.description}}};var v,x,y;o.parameters={...o.parameters,docs:{...(v=o.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    label: 'Email address',
    required: true,
    variant: 'outlined'
  }
}`,...(y=(x=o.parameters)==null?void 0:x.docs)==null?void 0:y.source}}};var E,q,S;n.parameters={...n.parameters,docs:{...(E=n.parameters)==null?void 0:E.docs,source:{originalSource:`{
  args: {
    required: true
  }
}`,...(S=(q=n.parameters)==null?void 0:q.docs)==null?void 0:S.source}}};var j,N,T;d.parameters={...d.parameters,docs:{...(j=d.parameters)==null?void 0:j.docs,source:{originalSource:`{
  args: {
    defaultValue: 'not-an-email',
    error: true,
    errorMessage: 'Enter a valid email address.',
    label: 'Email address',
    required: true,
    variant: 'outlined'
  }
}`,...(T=(N=d.parameters)==null?void 0:N.docs)==null?void 0:T.source}}};var I,L,W,B,F;t.parameters={...t.parameters,docs:{...(I=t.parameters)==null?void 0:I.docs,source:{originalSource:`{
  args: {
    disabled: true,
    variant: 'outlined'
  }
}`,...(W=(L=t.parameters)==null?void 0:L.docs)==null?void 0:W.source},description:{story:"Use the `disabled` prop to make the input non-interactive and appears faded,\nindicating that input is not currently accepted.",...(F=(B=t.parameters)==null?void 0:B.docs)==null?void 0:F.description}}};var w,D,R;l.parameters={...l.parameters,docs:{...(w=l.parameters)==null?void 0:w.docs,source:{originalSource:`{
  args: {
    id: 'email',
    label: 'Email'
  }
}`,...(R=(D=l.parameters)==null?void 0:D.docs)==null?void 0:R.source}}};var U,k,H,M,V;s.parameters={...s.parameters,docs:{...(U=s.parameters)==null?void 0:U.docs,source:{originalSource:`{
  args: {
    id: 'email-2',
    label: 'Email',
    helperText: 'Enter your email address.',
    required: true,
    variant: 'outlined'
  }
}`,...(H=(k=s.parameters)==null?void 0:k.docs)==null?void 0:H.source},description:{story:`Use the helperText prop to provide additional instructions
or information to users.`,...(V=(M=s.parameters)==null?void 0:M.docs)==null?void 0:V.description}}};var _,C,O,z,A;i.parameters={...i.parameters,docs:{...(_=i.parameters)==null?void 0:_.docs,source:{originalSource:`{
  render: args => <div className='flex items-center space-x-2'>
            <Input {...args} />
            <Button className='rounded-md'>Subscribe</Button>
        </div>
}`,...(O=(C=i.parameters)==null?void 0:C.docs)==null?void 0:O.source},description:{story:"Use the `Button` component to indicate that the input field can be submitted\nor used to trigger an action.",...(A=(z=i.parameters)==null?void 0:z.docs)==null?void 0:A.description}}};var G,J,K;c.parameters={...c.parameters,docs:{...(G=c.parameters)==null?void 0:G.docs,source:{originalSource:`{
  args: {
    type: 'file',
    placeholder: 'No file selected.'
  },
  render: args => <div className='grid w-full max-w-sm items-center gap-1.5'>
            <Label htmlFor='picture'>Upload Image</Label>
            <Input id='picture' type='file' {...args} variant={'outlined'} />
        </div>
}`,...(K=(J=c.parameters)==null?void 0:J.docs)==null?void 0:K.source}}};var P,Q,X;m.parameters={...m.parameters,docs:{...(P=m.parameters)==null?void 0:P.docs,source:{originalSource:`{
  args: {
    intent: 'time',
    type: 'time'
  },
  render: args => <div className='grid items-center gap-1.5'>
            <Label htmlFor='time'>Time</Label>
            <Input {...args} id='time' />
        </div>
}`,...(X=(Q=m.parameters)==null?void 0:Q.docs)==null?void 0:X.source}}};const he=["Default","Required","RequiredWithoutLabel","Error","Disabled","WithLabel","WithHelperText","WithButton","WithFile","TimeInput"];export{a as Default,t as Disabled,d as Error,o as Required,n as RequiredWithoutLabel,m as TimeInput,i as WithButton,c as WithFile,s as WithHelperText,l as WithLabel,he as __namedExportsOrder,ge as default};
