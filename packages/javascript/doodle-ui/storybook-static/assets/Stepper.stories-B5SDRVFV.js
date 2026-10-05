import{j as t}from"./jsx-runtime-BTJTZTIL.js";import{r as d}from"./index-2AqrBsEA.js";import{B as V}from"./Button-DTtPuz4n.js";import{T as F}from"./Typography-BMbK6Q6M.js";import{F as Ke,p as Ge}from"./index.es-C0EmWI2V.js";import{c as u}from"./utils-BIldDEfN.js";import{u as Je,m as Qe}from"./useRenderElement--ncz4fp5.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-Bb4qSo10.js";import"./Button.styles-DqX4pzJt.js";function Xe(e){return Je(e.defaultTagName??"div",e,e)}const _e=d.createContext(void 0),ze=d.createContext(void 0);function b(){const e=d.useContext(_e);if(!e)throw new Error("useStepper must be used within a Stepper");return e}function D(){const e=d.useContext(ze);if(!e)throw new Error("useStepItem must be used within a StepperItem");return e}const N=({steps:e,className:a,forceMount:n,panelHeader:r,navClassName:c,indicators:l={loading:t.jsx(Ke,{icon:Ge,className:"size-4 animate-spin","aria-label":"Loading"})},...p})=>t.jsxs(qe,{...p,indicators:l,className:u("flex min-h-0 flex-col gap-8",a),children:[t.jsx(Ve,{className:u("shrink-0",c),children:e.map((s,i)=>t.jsxs(Re,{step:i+1,isLoading:s.isLoading,isDisabled:s.isDisabled,isCompleted:s.isCompleted,children:[t.jsxs(Ye,{children:[t.jsx(Ee,{children:i+1}),s.description!=null?t.jsxs("div",{className:"flex flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center",children:[t.jsx(X,{children:s.title}),t.jsx(He,{children:s.description})]}):t.jsx(X,{children:s.title})]}),i<e.length-1&&t.jsx(Be,{})]},s.title))}),t.jsxs(Fe,{className:"min-h-0 flex-1 overflow-y-auto overscroll-contain",children:[r,e.map((s,i)=>t.jsx(Me,{value:i+1,forceMount:n,children:s.content},i))]})]}),qe=({defaultValue:e=1,value:a,onValueChange:n,orientation:r="horizontal",className:c,children:l,indicators:p={},...s})=>{var Y;const i=d.useId(),[x,S]=d.useState(e),[M,w]=d.useState(0),[O,j]=d.useState([]),h=d.useRef(new Map),$=d.useCallback((o,g)=>(h.current.set(o,g),w(h.current.size),()=>{h.current.delete(o),w(h.current.size)}),[]),U=d.useCallback((o,g,y)=>(j(C=>[...C,{node:o,step:g,isDisabled:y}]),()=>j(C=>C.filter(Q=>Q.node!==o))),[]),W=d.useCallback(o=>{!h.current.has(o)||h.current.get(o)||(a===void 0&&S(o),n==null||n(o))},[a,n]),I=a??x,m=()=>O.filter(({node:o,isDisabled:g})=>o.isConnected&&!o.disabled&&!g).sort((o,g)=>o.node.compareDocumentPosition(g.node)&Node.DOCUMENT_POSITION_FOLLOWING?-1:1),k=(o,g)=>{var Z;const y=m(),C=y.findIndex(({node:We})=>We===o);if(C===-1)return;const Q=(C+g+y.length)%y.length;(Z=y[Q])==null||Z.node.focus()},K=o=>k(o,1),P=o=>k(o,-1),G=()=>{var o;return(o=m()[0])==null?void 0:o.node.focus()},J=()=>{var o;return(o=m().at(-1))==null?void 0:o.node.focus()},f=m(),$e=(Y=f.find(({step:o})=>o===I)??f[0])==null?void 0:Y.step,Ue={id:i,tabStopStep:$e,activeStep:I,setActiveStep:W,stepsCount:M,orientation:r,registerStep:$,registerTrigger:U,focusNext:K,focusPrev:P,focusFirst:G,focusLast:J,indicators:p};return t.jsx(_e.Provider,{value:Ue,children:t.jsx("div",{"data-slot":"stepper",className:u("w-full font-sans text-text-main",c),"data-orientation":r,...s,children:l})})};function Re({step:e,isCompleted:a=!1,isDisabled:n=!1,isLoading:r=!1,className:c,children:l,...p}){const{activeStep:s,registerStep:i}=b();d.useLayoutEffect(()=>i(e,n),[i,e,n]);const x=s===e?"active":a||e<s?"completed":"inactive",S=r&&e===s;return t.jsx(ze.Provider,{value:{step:e,state:x,isDisabled:n,isLoading:S,isCompleted:a},children:t.jsx("div",{"data-slot":"stepper-item",className:u("group/step flex flex-auto items-center last:flex-none group-data-[orientation=horizontal]/stepper-nav:flex-row group-data-[orientation=vertical]/stepper-nav:flex-col",c),"data-state":x,...S?{"data-loading":!0}:{},...p,children:l})})}function Ye({className:e,children:a,tabIndex:n,render:r,disabled:c=!1,...l}){const{state:p,isLoading:s,step:i,isDisabled:x}=D(),{id:S,tabStopStep:M,setActiveStep:w,activeStep:O,registerTrigger:j,focusNext:h,focusPrev:$,focusFirst:U,focusLast:W}=b(),I=O===i,m=x||c,k=`${S}-tab-${i}`,K=`${S}-panel-${i}`,P=d.useRef(null);d.useLayoutEffect(()=>{if(P.current)return j(P.current,i,m)},[j,i,m]);const G=f=>{if(!m)switch(f.key){case"ArrowRight":case"ArrowDown":f.preventDefault(),h(f.currentTarget);break;case"ArrowLeft":case"ArrowUp":f.preventDefault(),$(f.currentTarget);break;case"Home":f.preventDefault(),U();break;case"End":f.preventDefault(),W();break;case"Enter":case" ":f.preventDefault(),w(i);break}},J={type:"button",role:"tab",id:k,"aria-selected":I,"aria-controls":K,"aria-disabled":m||void 0,tabIndex:m?-1:n??(M===i?0:-1),"data-slot":"stepper-trigger","data-state":p,"data-loading":s,className:u("inline-flex shrink-0 cursor-pointer items-center gap-1.5 rounded px-2 py-1 text-left outline-none focus-visible:z-10 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-secondary disabled:pointer-events-none disabled:opacity-60","max-sm:group-data-[orientation=horizontal]/stepper-nav:flex-col max-sm:group-data-[orientation=horizontal]/stepper-nav:text-center",e),onClick:()=>{m||w(i)},onKeyDown:G,disabled:m,children:a};return Xe({defaultTagName:"button",render:r,ref:P,props:Qe(J,l)})}function Ee({children:e,className:a,...n}){const{state:r,isLoading:c}=D(),{indicators:l}=b();return t.jsx("div",{...n,"data-slot":"stepper-indicator","data-state":r,className:u("group/indicator flex size-10 shrink-0 items-center justify-center rounded border border-transparent text-base font-bold leading-6 data-[state=active]:border-secondary",a),children:t.jsx("div",{className:u("flex size-8 items-center justify-center rounded border border-transparent group-data-[state=inactive]/indicator:border-[#939597] dark:group-data-[state=inactive]/indicator:border-input-border-default","group-data-[state=completed]/indicator:border-primary group-data-[state=completed]/indicator:bg-primary group-data-[state=completed]/indicator:text-text-contrast","group-data-[state=active]/indicator:rounded-sm group-data-[state=active]/indicator:border-secondary group-data-[state=active]/indicator:bg-secondary group-data-[state=active]/indicator:text-text-contrast"),children:c&&l.loading||l[r]||e})})}function Be({className:e,...a}){const{state:n}=D();return t.jsx("div",{...a,"data-slot":"stepper-separator","data-state":n,className:u("max-sm:group-data-[orientation=horizontal]/stepper-nav:mt-6 max-sm:group-data-[orientation=horizontal]/stepper-nav:self-start","bg-neutral-300 group-data-[orientation=horizontal]/stepper-nav:h-px group-data-[orientation=horizontal]/stepper-nav:min-w-4 group-data-[orientation=horizontal]/stepper-nav:flex-1 group-data-[orientation=vertical]/stepper-nav:h-12 group-data-[orientation=vertical]/stepper-nav:w-px",e)})}function X({children:e,className:a,...n}){const{state:r}=D();return t.jsx(F,{...n,variant:"h3","data-slot":"stepper-title","data-state":r,className:u("whitespace-nowrap text-base font-normal leading-4",a),children:e})}function He({children:e,className:a,...n}){const{state:r}=D();return t.jsx(F,{...n,component:"div","data-slot":"stepper-description",variant:"subtitle2","data-state":r,className:u("text-text-muted leading-4",a),children:e})}function Ve({children:e,className:a,...n}){const{activeStep:r,orientation:c}=b();return t.jsx("div",{...n,role:"tablist","aria-orientation":c,"data-slot":"stepper-nav","data-state":r,"data-orientation":c,className:u("group/stepper-nav flex items-center overflow-x-auto data-[orientation=horizontal]:w-full data-[orientation=horizontal]:flex-row data-[orientation=vertical]:flex-col max-sm:data-[orientation=horizontal]:items-start",a),children:e})}function Fe({children:e,className:a,...n}){const{activeStep:r}=b();return t.jsx("div",{...n,"data-slot":"stepper-panel","data-state":r,className:u("w-full",a),children:e})}function Me({value:e,forceMount:a,children:n,className:r,...c}){const{activeStep:l,id:p}=b(),s=e===l;return t.jsx("div",{tabIndex:0,...c,id:`${p}-panel-${e}`,role:"tabpanel","aria-labelledby":`${p}-tab-${e}`,"data-slot":"stepper-content","data-state":l,className:u("w-full",r,!s&&"hidden"),hidden:!s,children:(s||a)&&n})}N.__docgenInfo={description:"A preset layout for simple, ordered flows. Use Stepper composition for custom layouts.",methods:[],displayName:"BasicStepper",props:{steps:{required:!0,tsType:{name:"Array",elements:[{name:"Step"}],raw:"Step[]"},description:""},forceMount:{required:!1,tsType:{name:"boolean"},description:"Keep inactive content mounted, for example to retain form registration and validation."},panelHeader:{required:!1,tsType:{name:"ReactNode"},description:"Shared content above the step panels, such as a form summary."},navClassName:{required:!1,tsType:{name:"string"},description:"Additional navigation styling, for example clearance for a drawer's close button."},indicators:{defaultValue:{value:`{
    loading: <FontAwesomeIcon icon={faSpinner} className='size-4 animate-spin' aria-label='Loading' />,
}`,computed:!1},required:!1}},composes:["Omit"]};qe.__docgenInfo={description:"",methods:[],displayName:"Stepper",props:{defaultValue:{required:!1,tsType:{name:"number"},description:"Initial active step (1-based) for an uncontrolled stepper.",defaultValue:{value:"1",computed:!1}},value:{required:!1,tsType:{name:"number"},description:"Active step (1-based). The parent must enable the destination before setting this value."},onValueChange:{required:!1,tsType:{name:"signature",type:"function",raw:"(value: number) => void",signature:{arguments:[{type:{name:"number"},name:"value"}],return:{name:"void"}}},description:""},orientation:{required:!1,tsType:{name:"union",raw:"'horizontal' | 'vertical'",elements:[{name:"literal",value:"'horizontal'"},{name:"literal",value:"'vertical'"}]},description:"",defaultValue:{value:"'horizontal'",computed:!1}},indicators:{required:!1,tsType:{name:"Partial",elements:[{name:"Record",elements:[{name:"union",raw:"'active' | 'completed' | 'inactive' | 'loading'",elements:[{name:"literal",value:"'active'"},{name:"literal",value:"'completed'"},{name:"literal",value:"'inactive'"},{name:"literal",value:"'loading'"}]},{name:"ReactNode"}],raw:"Record<StepState, ReactNode>"}],raw:"Partial<Record<StepState, ReactNode>>"},description:"",defaultValue:{value:"{}",computed:!1}}},composes:["HTMLAttributes"]};Me.__docgenInfo={description:"",methods:[],displayName:"StepperContent",props:{value:{required:!0,tsType:{name:"number"},description:""},forceMount:{required:!1,tsType:{name:"boolean"},description:""}},composes:["ComponentProps"]};He.__docgenInfo={description:"",methods:[],displayName:"StepperDescription"};Ee.__docgenInfo={description:"",methods:[],displayName:"StepperIndicator"};Re.__docgenInfo={description:"",methods:[],displayName:"StepperItem",props:{step:{required:!0,tsType:{name:"number"},description:""},isCompleted:{required:!1,tsType:{name:"boolean"},description:"",defaultValue:{value:"false",computed:!1}},isDisabled:{required:!1,tsType:{name:"boolean"},description:"",defaultValue:{value:"false",computed:!1}},isLoading:{required:!1,tsType:{name:"boolean"},description:"",defaultValue:{value:"false",computed:!1}}},composes:["HTMLAttributes"]};Ve.__docgenInfo={description:"",methods:[],displayName:"StepperNav"};Fe.__docgenInfo={description:"",methods:[],displayName:"StepperPanel"};Be.__docgenInfo={description:"",methods:[],displayName:"StepperSeparator"};X.__docgenInfo={description:"",methods:[],displayName:"StepperTitle"};const ft={title:"Components/Stepper",component:N,parameters:{layout:"padded",docs:{description:{component:'BasicStepper is a preset for simple, ordered flows. Steps can include an optional description below the title. Use value/onValueChange to share state with controls outside the stepper, or useStepper() in step content for local navigation. Use the composable Stepper primitives for custom layouts, per-step indicators, and richer workflows. Disabled steps reject clicks and useStepper navigation requests without calling onValueChange. Parent controls must enable the destination before setting value; update both together after validation succeeds. Values are 1-based. Inactive panels stay hidden while their children unmount, unless forceMount is enabled. Content uses normal document layout. In a container with a constrained height, use className="flex-1" to keep navigation visible while the panel scrolls.'}}},argTypes:{forceMount:{description:"Keep inactive step content mounted and hidden to retain form registration and validation.",control:"boolean",table:{defaultValue:{summary:"false"}}},panelHeader:{description:"Shared content above the step panels that remains mounted during navigation.",control:!1},navClassName:{description:"Additional classes on the navigation, such as clearance for a drawer close button.",control:"text"},steps:{description:"Step titles, optional descriptions, content, loading state, and disabled state. Enable disabled steps before navigating to them.",control:{type:"object"}},defaultValue:{description:"The default step to be selected (1-based index).",control:{type:"number"},table:{type:{summary:"number"},defaultValue:{summary:"1"}}},value:{description:"The controlled active step (1-based). The parent accepts navigation requests via onValueChange and must enable the destination before selecting it.",control:{type:"number"},table:{type:{summary:"number"}}},onValueChange:{description:"Called when a trigger or useStepper().setActiveStep requests an enabled step.",control:!1,table:{type:{summary:"(nextStep: number) => void"}}},indicators:{table:{type:{summary:"StepIndicators"}}}},decorators:[e=>t.jsx("div",{className:"mx-auto w-full max-w-[796px]",children:t.jsx(e,{})})],tags:["autodocs"]},v={args:{steps:[{title:"Profile",content:"Profile content"},{title:"Secret",content:"Secret Content"},{title:"Schedule",content:"Schedule Content"}]}},A={args:{steps:[{title:"Profile",description:"Name and details",content:"Profile content"},{title:"Secret",description:"Access credentials",content:"Secret content"},{title:"Schedule",content:"Schedule content"}]}};function Ze(e){return t.jsx(N,{...e,panelHeader:t.jsxs("div",{className:"mb-6 flex flex-col gap-2",children:[t.jsx(F,{variant:"h2",children:"Create a collection plan"}),t.jsx(F,{variant:"body2",children:"Configure the profile, credentials, and schedule for your collection."})]})})}const L={args:v.args,render:e=>t.jsx(Ze,{...e}),parameters:{docs:{description:{story:"Use panelHeader for a shared title, description, or summary above the step content. Select another step to see the header stay in place while the content changes."},source:{code:`import { BasicStepper, Typography } from 'doodle-ui';

<BasicStepper
    steps={[
        { title: 'Profile', content: 'Profile content' },
        { title: 'Secret', content: 'Secret content' },
        { title: 'Schedule', content: 'Schedule content' },
    ]}
    panelHeader={
        <div className='mb-6 flex flex-col gap-2'>
            <Typography variant='h2'>Create a collection plan</Typography>
            <Typography variant='body2'>
                Configure the profile, credentials, and schedule for your collection.
            </Typography>
        </div>
    }
/>`}}}},_={args:{steps:[{title:"Account",content:"Account Content"},{title:"Profile",content:"Profile Content"},{title:"Preferences",content:"Preferences Content"},{title:"Review",content:"Review Content"},{title:"Complete",content:"Complete Content"}]}};function et({defaultValue:e=1,onValueChange:a,...n}){var l,p;const[r,c]=d.useState(e);return t.jsxs("div",{className:"flex flex-col gap-6",children:[t.jsx(N,{...n,value:r,onValueChange:s=>{c(s),a==null||a(s)}}),t.jsxs("div",{className:"flex items-center gap-4",children:[t.jsx(V,{disabled:r<=1||((l=n.steps[r-2])==null?void 0:l.isDisabled),onClick:()=>c(r-1),children:"Back"}),t.jsx(V,{disabled:r>=n.steps.length||((p=n.steps[r])==null?void 0:p.isDisabled),onClick:()=>c(r+1),children:"Next"}),t.jsxs("p",{role:"status",children:["Step ",r," of ",n.steps.length]})]})]})}const z={args:v.args,render:e=>t.jsx(et,{...e}),parameters:{docs:{description:{story:"Back and Next sit outside BasicStepper and share the active step through controlled props."}}}};function Oe({title:e,isLast:a}){const{activeStep:n,setActiveStep:r}=b();return t.jsxs("div",{className:"flex flex-col gap-4 text-center items-center",children:[t.jsxs("p",{children:[e," content"]}),!a&&t.jsx(V,{onClick:()=>r(n+1),children:"Continue"})]})}function tt(e){return t.jsx(N,{...e,steps:e.steps.map((a,n)=>({...a,content:t.jsx(Oe,{title:a.title,isLast:n===e.steps.length-1})}))})}const q={args:v.args,render:e=>t.jsx(tt,{...e}),parameters:{docs:{description:{story:"Continue is inside step content and uses the existing useStepper hook. No parent state is needed."}}}},T={args:{steps:[{title:"Profile",content:"Profile content"},{title:"Secret",content:"Secret content",isDisabled:!0},{title:"Schedule",content:"Schedule content"}]},parameters:{docs:{description:{story:"Secret is unavailable to clicks, keyboard activation, and useStepper navigation requests. Arrow keys skip it. Enable it before navigating to it."}}}},R={args:{steps:[{title:"Profile",content:"Profile content",isDisabled:!0},{title:"Secret",content:"Secret content"},{title:"Schedule",content:"Schedule content"}]},parameters:{docs:{description:{story:"If the selected step is disabled, Tab reaches the first enabled trigger. Selection stays unchanged until an enabled step is activated."}}}};function nt({steps:e,onValueChange:a,...n}){const[r,c]=d.useState(1),[l,p]=d.useState(!1);return t.jsxs("div",{className:"flex flex-col gap-6 items-start",children:[t.jsx(N,{...n,value:r,onValueChange:s=>{c(s),a==null||a(s)},steps:e.map((s,i)=>({...s,isDisabled:i===1?!l:s.isDisabled,content:i===0?t.jsx(Oe,{title:s.title,isLast:!1}):s.content}))}),t.jsx("p",{role:"status",children:l?"Secret is available.":"Complete Profile to unlock Secret. Continue is blocked until then."}),t.jsx(V,{disabled:l,onClick:()=>{p(!0),c(2)},children:"Complete Profile"})]})}const E={args:T.args,render:e=>t.jsx(nt,{...e}),parameters:{docs:{description:{story:"Continue uses useStepper and cannot enter disabled Secret. Complete Profile enables Secret and sets the controlled value in the same update, as a parent would after successful validation."}}}},B={args:{steps:[{title:"Profile",content:"Profile content"},{title:"Secret",content:"Secret Content",isLoading:!0},{title:"Schedule",content:"Schedule Content",isDisabled:!0}],defaultValue:2}},H={args:{steps:[{title:"Profile",content:"Profile content"},{title:"Secret",content:"Secret content"},{title:"Schedule",content:"Schedule content",isCompleted:!0}]},parameters:{docs:{description:{story:"Completed steps are visually distinct but can still be selected."}}}};var ee,te,ne;v.parameters={...v.parameters,docs:{...(ee=v.parameters)==null?void 0:ee.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      content: 'Profile content'
    }, {
      title: 'Secret',
      content: 'Secret Content'
    }, {
      title: 'Schedule',
      content: 'Schedule Content'
    }]
  }
}`,...(ne=(te=v.parameters)==null?void 0:te.docs)==null?void 0:ne.source}}};var ae,re,se;A.parameters={...A.parameters,docs:{...(ae=A.parameters)==null?void 0:ae.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      description: 'Name and details',
      content: 'Profile content'
    }, {
      title: 'Secret',
      description: 'Access credentials',
      content: 'Secret content'
    }, {
      title: 'Schedule',
      content: 'Schedule content'
    }]
  }
}`,...(se=(re=A.parameters)==null?void 0:re.docs)==null?void 0:se.source}}};var oe,ie,ce;L.parameters={...L.parameters,docs:{...(oe=L.parameters)==null?void 0:oe.docs,source:{originalSource:`{
  args: Default.args,
  render: args => <PanelHeaderStepper {...args} />,
  parameters: {
    docs: {
      description: {
        story: 'Use panelHeader for a shared title, description, or summary above the step content. Select another step to see the header stay in place while the content changes.'
      },
      source: {
        code: \`import { BasicStepper, Typography } from 'doodle-ui';

<BasicStepper
    steps={[
        { title: 'Profile', content: 'Profile content' },
        { title: 'Secret', content: 'Secret content' },
        { title: 'Schedule', content: 'Schedule content' },
    ]}
    panelHeader={
        <div className='mb-6 flex flex-col gap-2'>
            <Typography variant='h2'>Create a collection plan</Typography>
            <Typography variant='body2'>
                Configure the profile, credentials, and schedule for your collection.
            </Typography>
        </div>
    }
/>\`
      }
    }
  }
}`,...(ce=(ie=L.parameters)==null?void 0:ie.docs)==null?void 0:ce.source}}};var le,de,pe;_.parameters={..._.parameters,docs:{...(le=_.parameters)==null?void 0:le.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Account',
      content: 'Account Content'
    }, {
      title: 'Profile',
      content: 'Profile Content'
    }, {
      title: 'Preferences',
      content: 'Preferences Content'
    }, {
      title: 'Review',
      content: 'Review Content'
    }, {
      title: 'Complete',
      content: 'Complete Content'
    }]
  }
}`,...(pe=(de=_.parameters)==null?void 0:de.docs)==null?void 0:pe.source}}};var ue,me,fe;z.parameters={...z.parameters,docs:{...(ue=z.parameters)==null?void 0:ue.docs,source:{originalSource:`{
  args: Default.args,
  render: args => <ControlledStepper {...args} />,
  parameters: {
    docs: {
      description: {
        story: 'Back and Next sit outside BasicStepper and share the active step through controlled props.'
      }
    }
  }
}`,...(fe=(me=z.parameters)==null?void 0:me.docs)==null?void 0:fe.source}}};var ge,he,Se;q.parameters={...q.parameters,docs:{...(ge=q.parameters)==null?void 0:ge.docs,source:{originalSource:`{
  args: Default.args,
  render: args => <ContentNavigationStepper {...args} />,
  parameters: {
    docs: {
      description: {
        story: 'Continue is inside step content and uses the existing useStepper hook. No parent state is needed.'
      }
    }
  }
}`,...(Se=(he=q.parameters)==null?void 0:he.docs)==null?void 0:Se.source}}};var ve,be,xe;T.parameters={...T.parameters,docs:{...(ve=T.parameters)==null?void 0:ve.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      content: 'Profile content'
    }, {
      title: 'Secret',
      content: 'Secret content',
      isDisabled: true
    }, {
      title: 'Schedule',
      content: 'Schedule content'
    }]
  },
  parameters: {
    docs: {
      description: {
        story: 'Secret is unavailable to clicks, keyboard activation, and useStepper navigation requests. Arrow keys skip it. Enable it before navigating to it.'
      }
    }
  }
}`,...(xe=(be=T.parameters)==null?void 0:be.docs)==null?void 0:xe.source}}};var ye,Ce,Ne;R.parameters={...R.parameters,docs:{...(ye=R.parameters)==null?void 0:ye.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      content: 'Profile content',
      isDisabled: true
    }, {
      title: 'Secret',
      content: 'Secret content'
    }, {
      title: 'Schedule',
      content: 'Schedule content'
    }]
  },
  parameters: {
    docs: {
      description: {
        story: 'If the selected step is disabled, Tab reaches the first enabled trigger. Selection stays unchanged until an enabled step is activated.'
      }
    }
  }
}`,...(Ne=(Ce=R.parameters)==null?void 0:Ce.docs)==null?void 0:Ne.source}}};var we,je,Pe;E.parameters={...E.parameters,docs:{...(we=E.parameters)==null?void 0:we.docs,source:{originalSource:`{
  args: DisabledStep.args,
  render: args => <DisabledNavigationStepper {...args} />,
  parameters: {
    docs: {
      description: {
        story: 'Continue uses useStepper and cannot enter disabled Secret. Complete Profile enables Secret and sets the controlled value in the same update, as a parent would after successful validation.'
      }
    }
  }
}`,...(Pe=(je=E.parameters)==null?void 0:je.docs)==null?void 0:Pe.source}}};var Te,De,Ie;B.parameters={...B.parameters,docs:{...(Te=B.parameters)==null?void 0:Te.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      content: 'Profile content'
    }, {
      title: 'Secret',
      content: 'Secret Content',
      isLoading: true
    }, {
      title: 'Schedule',
      content: 'Schedule Content',
      isDisabled: true
    }],
    defaultValue: 2
  }
}`,...(Ie=(De=B.parameters)==null?void 0:De.docs)==null?void 0:Ie.source}}};var ke,Ae,Le;H.parameters={...H.parameters,docs:{...(ke=H.parameters)==null?void 0:ke.docs,source:{originalSource:`{
  args: {
    steps: [{
      title: 'Profile',
      content: 'Profile content'
    }, {
      title: 'Secret',
      content: 'Secret content'
    }, {
      title: 'Schedule',
      content: 'Schedule content',
      isCompleted: true
    }]
  },
  parameters: {
    docs: {
      description: {
        story: 'Completed steps are visually distinct but can still be selected.'
      }
    }
  }
}`,...(Le=(Ae=H.parameters)==null?void 0:Ae.docs)==null?void 0:Le.source}}};const gt=["Default","WithDescriptions","WithPanelHeader","LongerFlow","Controlled","ContentNavigation","DisabledStep","DisabledInitialStep","DisabledNavigation","Loading","CompletedSteps"];export{H as CompletedSteps,q as ContentNavigation,z as Controlled,v as Default,R as DisabledInitialStep,E as DisabledNavigation,T as DisabledStep,B as Loading,_ as LongerFlow,A as WithDescriptions,L as WithPanelHeader,gt as __namedExportsOrder,ft as default};
