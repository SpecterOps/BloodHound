import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{F as u,u as h,a as m,b as t,c as s,d as a,e as i,f as c}from"./Form-DhwQ8OSg.js";import{B as f}from"./Button-DTtPuz4n.js";import{I as x}from"./Input--sj12WP2.js";import{S as b}from"./Switch-B-7ih0xl.js";import"./index-ChsGqxH_.js";import"./index-TwK_z8r1.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./Label-Dsc1dydc.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./Typography-BMbK6Q6M.js";import"./index-Dv0R_RYs.js";import"./index-Dc_EP5I0.js";import"./index-CFknkR0k.js";import"./index-DRLFypRQ.js";const H={title:"Components/Form",component:u,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{},args:{}},n={render:()=>{const o=h({defaultValues:{username:"doodle",checked:!0}}),F=r=>{alert(JSON.stringify(r))};return e.jsx(u,{...o,children:e.jsxs("form",{onSubmit:o.handleSubmit(F),className:"space-y-8",children:[e.jsx(m,{control:o.control,name:"username",render:({field:r})=>e.jsxs(t,{children:[e.jsx(s,{children:"Username"}),e.jsx(a,{children:e.jsx(x,{placeholder:"Username",...r})}),e.jsx(i,{children:"This is your public display name."}),e.jsx(c,{})]})}),e.jsx(m,{control:o.control,name:"checked",render:({field:r})=>e.jsxs(t,{children:[e.jsx(s,{children:"Toggle me!"}),e.jsx(a,{children:e.jsx(b,{checked:r.value,onCheckedChange:r.onChange})}),e.jsx(i,{children:"Switch it up"}),e.jsx(c,{})]})}),e.jsx(f,{type:"submit",children:"Submit"})]})})}};var l,d,p;n.parameters={...n.parameters,docs:{...(l=n.parameters)==null?void 0:l.docs,source:{originalSource:`{
  render: () => {
    const form = useForm<FormInputs>({
      defaultValues: {
        username: 'doodle',
        checked: true
      }
    });
    const onSubmit: SubmitHandler<FormInputs> = data => {
      alert(JSON.stringify(data));
    };
    return <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-8'>
                    <FormField control={form.control} name='username' render={({
          field
        }) => <FormItem>
                                <FormLabel>Username</FormLabel>
                                <FormControl>
                                    <Input placeholder='Username' {...field} />
                                </FormControl>
                                <FormDescription>This is your public display name.</FormDescription>
                                <FormMessage />
                            </FormItem>} />
                    <FormField control={form.control} name='checked' render={({
          field
        }) => <FormItem>
                                <FormLabel>Toggle me!</FormLabel>
                                <FormControl>
                                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                                </FormControl>
                                <FormDescription>Switch it up</FormDescription>
                                <FormMessage />
                            </FormItem>} />
                    <Button type='submit'>Submit</Button>
                </form>
            </Form>;
  }
}`,...(p=(d=n.parameters)==null?void 0:d.docs)==null?void 0:p.source}}};const R=["Default"];export{n as Default,R as __namedExportsOrder,H as default};
