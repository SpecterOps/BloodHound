import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{r as N}from"./index-2AqrBsEA.js";import{B as I}from"./Button-DTtPuz4n.js";import{F as E,m as R}from"./index.es-C0EmWI2V.js";import{c as T}from"./index-Bb4qSo10.js";import{A as n}from"./AppIcons-Dp2sojKC.js";import{c as p}from"./utils-BIldDEfN.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./Card-BdBAz3fL.js";import"./Typography-BMbK6Q6M.js";import"./UserCog-VADywgBc.js";import"./VisuallyHidden-DT_8ZbEK.js";import"./index-oIeaYTeP.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";const W=T("flex text-sm w-full rounded px-4 py-2 ",{variants:{variant:{default:"bg-status-indeterminate-fill text-neutral-1",error:"bg-status-error-fill text-status-error-text [&>div>svg]:text-status-error-main",info:"bg-status-info-fill text-status-info-text [&>div>svg]:text-status-info-main",success:"bg-status-success-fill text-status-success-text [&>div>svg]:text-status-success-main",warning:"bg-status-warning-fill text-status-warning-text [&>div>svg]:text-status-warning-main"}},defaultVariants:{variant:"default"}}),P={default:e.jsx(n.Info,{}),error:e.jsx(n.Error,{}),info:e.jsx(n.Info,{}),success:e.jsx(n.CircleCheck,{}),warning:e.jsx(n.Warning,{})},t=N.forwardRef(({action:s,children:c,className:i,onClose:u,title:d,variant:m="default",...S},V)=>e.jsxs("div",{ref:V,role:"alert",className:p(W({variant:m}),i),...S,children:[e.jsx("div",{className:"my-3 mr-2",children:P[m??"default"]}),e.jsxs("div",{className:"flex flex-col justify-center flex-grow py-2 gap-1",children:[d&&e.jsx("h5",{className:p("text-base/6 font-medium tracking-tight"),children:d}),c]}),e.jsxs("div",{className:"flex mt-1",children:[s&&e.jsx("button",{"aria-label":"Alert action",className:"h-6 rounded-sm px-1 font-medium uppercase focus:outline-none focus-visible:focus-ring",onClick:s.onClick,type:"button",children:s.label}),u&&e.jsx("button",{"aria-label":"Dismiss alert",className:"ml-2 inline-flex h-6 items-center justify-center rounded-sm px-1 focus:outline-none focus-visible:focus-ring",onClick:u,type:"button",children:e.jsx(E,{icon:R,className:"mt-1 size-4"})})]})]}));t.displayName="Alert";t.__docgenInfo={description:"",methods:[],displayName:"Alert",props:{action:{required:!1,tsType:{name:"signature",type:"object",raw:"{ label: string; onClick: () => void }",signature:{properties:[{key:"label",value:{name:"string",required:!0}},{key:"onClick",value:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}},required:!0}}]}},description:""},children:{required:!0,tsType:{name:"ReactNode"},description:""},onClose:{required:!1,tsType:{name:"signature",type:"function",raw:"() => void",signature:{arguments:[],return:{name:"void"}}},description:""},title:{required:!1,tsType:{name:"string"},description:""},variant:{defaultValue:{value:"'default'",computed:!1},required:!1}}};const ee={title:"Components/Alert",component:t,parameters:{layout:"centered",docs:{description:{component:"A contextual feedback message for user actions. Alerts communicate status or important information and support an optional title alongside body content. Use the `variant` prop to convey different severities — `default`, `error`, `info`, `success`, or `warning`."}}},tags:["autodocs"],argTypes:{children:{control:"text",description:"The body content of the alert."},action:{control:"object",description:"Optional action button rendered inside the alert. Accepts an object with a `label` string and an `onClick` handler.",table:{type:{summary:"{ label: string; onClick: () => void }"},defaultValue:{summary:"undefined"}}},onClose:{control:!1,description:"Optional callback invoked when the close button is clicked. When provided, a close button is rendered in the alert. When omitted, no close button is shown.",table:{type:{summary:"() => void"},defaultValue:{summary:"undefined"}}},title:{control:"text",description:"Optional heading rendered above the alert body content."},variant:{control:"select",options:["default","error","info","success","warning"],description:"Controls the visual style of the alert to convey semantic meaning.",table:{defaultValue:{summary:"default"}}},className:{control:"text",description:"Additional CSS class names to apply to the alert container."}},args:{children:"This is an alert message."}},r={args:{children:"This is a default alert",title:"Heads up!"}},a={parameters:{docs:{description:{story:"Alerts do not manage their own visibility state, but when `onClose` is provided, a close button appears in the alert. Clicking it invokes the callback — use this to remove the alert from the UI or update state in the parent component."},source:{code:`
const [visible, setVisible] = useState(true);

return (
    <div className='flex flex-col items-center gap-4 w-[400px]'>
        {visible ? (
            <Alert title="I'm dismissible" onClose={() => setVisible(false)}>
                Click the close button to dismiss this alert.
            </Alert>
        ) : (
            <Button onClick={() => setVisible(true)} type='button'>
                Reset alert
            </Button>
        )}
    </div>
);`,language:"tsx",type:"code"}}},render:function(){const[c,i]=N.useState(!0);return e.jsx("div",{className:"flex flex-col items-center gap-4 w-[400px]",children:c?e.jsx(t,{title:"I'm dismissible",onClose:()=>i(!1),children:"Click the close button to dismiss this alert."}):e.jsx(I,{onClick:()=>i(!0),type:"button",children:"Reset alert"})})}},o={args:{action:{label:"Alert",onClick:()=>alert("You clicked it!")},children:"Click 'ALERT' to call action.",title:"With Action"},parameters:{docs:{description:{story:"Use the `action` prop to render a labelled button inside the alert. Provide a `label` string and an `onClick` handler — this is useful for surfacing a single contextual action alongside the alert message, such as a retry, undo, or navigation link."}}}},l={parameters:{docs:{description:{story:"Each `variant` applies a distinct visual style to communicate semantic meaning. Use `default` for general messages, `info` for informational context, `success` to confirm a completed action, `warning` to signal potential issues, and `error` to indicate something went wrong."},source:{code:`
<div className='flex flex-col gap-4 w-[480px]'>
    <Alert variant='default' title='Default'>
        A general-purpose alert with no specific severity.
    </Alert>
    <Alert variant='info' title='Info'>
        Here is some useful information for you.
    </Alert>
    <Alert variant='success' title='Success'>
        Your action was completed successfully.
    </Alert>    
    <Alert variant='warning' title='Warning'>
        Proceed with caution — this action may have unintended effects.
    </Alert>
    <Alert variant='error' title='Error'>
        Something went wrong. Please try again.
    </Alert>
</div>`,language:"tsx",type:"code"}}},render:()=>e.jsxs("div",{className:"flex flex-col gap-4 w-[480px]",children:[e.jsx(t,{variant:"default",title:"Default",children:"A general-purpose alert with no specific severity."}),e.jsx(t,{variant:"info",title:"Info",children:"Here is some useful information for you."}),e.jsx(t,{variant:"success",title:"Success",children:"Your action was completed successfully."}),e.jsx(t,{variant:"warning",title:"Warning",children:"Proceed with caution — this action may have unintended effects."}),e.jsx(t,{variant:"error",title:"Error",children:"Something went wrong. Please try again."})]})};var f,g,h;r.parameters={...r.parameters,docs:{...(f=r.parameters)==null?void 0:f.docs,source:{originalSource:`{
  args: {
    children: 'This is a default alert',
    title: 'Heads up!'
  }
}`,...(h=(g=r.parameters)==null?void 0:g.docs)==null?void 0:h.source}}};var v,x,b;a.parameters={...a.parameters,docs:{...(v=a.parameters)==null?void 0:v.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'Alerts do not manage their own visibility state, but when \`onClose\` is provided, a close button appears in the alert. Clicking it invokes the callback — use this to remove the alert from the UI or update state in the parent component.'
      },
      source: {
        code: \`
const [visible, setVisible] = useState(true);

return (
    <div className='flex flex-col items-center gap-4 w-[400px]'>
        {visible ? (
            <Alert title="I'm dismissible" onClose={() => setVisible(false)}>
                Click the close button to dismiss this alert.
            </Alert>
        ) : (
            <Button onClick={() => setVisible(true)} type='button'>
                Reset alert
            </Button>
        )}
    </div>
);\`,
        language: 'tsx',
        type: 'code'
      }
    }
  },
  render: function Render() {
    const [visible, setVisible] = useState(true);
    return <div className='flex flex-col items-center gap-4 w-[400px]'>
                {visible ? <Alert title="I'm dismissible" onClose={() => setVisible(false)}>
                        Click the close button to dismiss this alert.
                    </Alert> : <Button onClick={() => setVisible(true)} type='button'>
                        Reset alert
                    </Button>}
            </div>;
  }
}`,...(b=(x=a.parameters)==null?void 0:x.docs)==null?void 0:b.source}}};var y,A,w;o.parameters={...o.parameters,docs:{...(y=o.parameters)==null?void 0:y.docs,source:{originalSource:`{
  args: {
    action: {
      label: 'Alert',
      onClick: () => alert('You clicked it!')
    },
    children: "Click 'ALERT' to call action.",
    title: 'With Action'
  },
  parameters: {
    docs: {
      description: {
        story: 'Use the \`action\` prop to render a labelled button inside the alert. Provide a \`label\` string and an \`onClick\` handler — this is useful for surfacing a single contextual action alongside the alert message, such as a retry, undo, or navigation link.'
      }
    }
  }
}`,...(w=(A=o.parameters)==null?void 0:A.docs)==null?void 0:w.source}}};var k,C,j;l.parameters={...l.parameters,docs:{...(k=l.parameters)==null?void 0:k.docs,source:{originalSource:`{
  parameters: {
    docs: {
      description: {
        story: 'Each \`variant\` applies a distinct visual style to communicate semantic meaning. Use \`default\` for general messages, \`info\` for informational context, \`success\` to confirm a completed action, \`warning\` to signal potential issues, and \`error\` to indicate something went wrong.'
      },
      source: {
        code: \`
<div className='flex flex-col gap-4 w-[480px]'>
    <Alert variant='default' title='Default'>
        A general-purpose alert with no specific severity.
    </Alert>
    <Alert variant='info' title='Info'>
        Here is some useful information for you.
    </Alert>
    <Alert variant='success' title='Success'>
        Your action was completed successfully.
    </Alert>    
    <Alert variant='warning' title='Warning'>
        Proceed with caution — this action may have unintended effects.
    </Alert>
    <Alert variant='error' title='Error'>
        Something went wrong. Please try again.
    </Alert>
</div>\`,
        language: 'tsx',
        type: 'code'
      }
    }
  },
  render: () => <div className='flex flex-col gap-4 w-[480px]'>
            <Alert variant='default' title='Default'>
                A general-purpose alert with no specific severity.
            </Alert>
            <Alert variant='info' title='Info'>
                Here is some useful information for you.
            </Alert>
            <Alert variant='success' title='Success'>
                Your action was completed successfully.
            </Alert>
            <Alert variant='warning' title='Warning'>
                Proceed with caution — this action may have unintended effects.
            </Alert>
            <Alert variant='error' title='Error'>
                Something went wrong. Please try again.
            </Alert>
        </div>
}`,...(j=(C=l.parameters)==null?void 0:C.docs)==null?void 0:j.source}}};const te=["Default","Dismissible","WithAction","Variants"];export{r as Default,a as Dismissible,l as Variants,o as WithAction,te as __namedExportsOrder,ee as default};
