
import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { authService } from '@/services/authService';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { toast } from '@/components/ui/use-toast';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Eye, EyeOff, Mail, LogIn, AlertCircle, Github } from "lucide-react";
import { useLanguage } from '@/contexts/LanguageContext';
import { useTheme } from '@/contexts/ThemeContext';
import { ForgotPasswordDialog } from '@/components/auth/ForgotPasswordDialog';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showForgotPassword, setShowForgotPassword] = useState(false);
  const [loginError, setLoginError] = useState('');
  const navigate = useNavigate();
  const { t } = useLanguage();
  const { theme } = useTheme();

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setLoginError(''); // Clear previous errors

    try {
      await authService.login({ email, password });
      toast({
        title: t("loginSuccessful"),
        description: t("loginSuccessMessage"),
      });
      navigate('/dashboard');
    } catch (error) {
      console.error("Login error details:", error);
      
      // Set specific error message based on the error
      if (error instanceof Error) {
        if (error.message.includes('Failed to authenticate') || 
            error.message.includes('Authentication failed') ||
            error.message.includes('Invalid credentials') ||
            error.message.includes('invalid email or password')) {
          setLoginError(t("invalidCredentials"));
        } else {
          setLoginError(error.message);
        }
      } else {
        setLoginError(t("authenticationFailed"));
      }
      
      toast({
        variant: "destructive",
        title: t("loginFailed"),
        description: error instanceof Error 
          ? (error.message.includes('Failed to authenticate') || 
             error.message.includes('Authentication failed') ||
             error.message.includes('Invalid credentials') ||
             error.message.includes('invalid email or password')) 
            ? t("invalidCredentials")
            : error.message
          : t("authenticationFailed"),
      });
    } finally {
      setLoading(false);
    }
  }, [email, password, navigate, t]);

  const togglePasswordVisibility = useCallback(() => {
    setShowPassword(prev => !prev);
  }, []);

  const handleEmailChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    setLoginError(''); // Clear error when user starts typing
  }, []);

  const handlePasswordChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    setLoginError(''); // Clear error when user starts typing
  }, []);

  const openExternalLink = useCallback((url: string) => {
    window.open(url, '_blank');
  }, []);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background p-4">
      <div className="w-full max-w-md p-4 sm:p-8 space-y-6 rounded-xl bg-card shadow-xl border border-border/20">
        <div className="text-center">
          {/* Logo */}
          <div className="mb-4">
            <img 
              src="/checkcle_logo.svg" 
              alt="CheckCle Logo" 
              className="mx-auto h-16 w-auto"
            />
          </div>
          
          <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">{t("signInToYourAccount")}</h1>
          <span className="inline-block mt-2 text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">Modified Build</span>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 sm:space-y-6">
          {/* Error Alert */}
          {loginError && (
            <Alert variant="destructive" className="mb-4">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{loginError}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-1 sm:space-y-2">
            <label className="text-xs sm:text-sm font-medium text-foreground" htmlFor="email">{t("email")}</label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <Mail className="h-4 w-4 text-muted-foreground" />
              </div>
              <Input
                id="email"
                placeholder="your.email@provider.com"
                type="email"
                value={email}
                onChange={handleEmailChange}
                required
                className="pl-10 text-sm sm:text-base h-9 sm:h-10"
              />
            </div>
          </div>
          
          <div className="space-y-1 sm:space-y-2">
            <div className="flex items-center justify-between">
              <label className="text-xs sm:text-sm font-medium text-foreground" htmlFor="password">{t("password")}</label>
              <button 
                type="button"
                onClick={() => setShowForgotPassword(true)}
                className="text-xs sm:text-sm text-emerald-500 hover:text-emerald-400"
              >
                {t("forgot")}
              </button>
            </div>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-muted-foreground">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                </svg>
              </div>
              <Input
                id="password"
                placeholder="••••••••••••"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={handlePasswordChange}
                required
                className="pl-10 text-sm sm:text-base h-9 sm:h-10"
              />
              <button
                type="button"
                className="absolute inset-y-0 right-0 flex items-center pr-3"
                onClick={togglePasswordVisibility}
                aria-label={showPassword ? "Hide password" : "Show password"}
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <Eye className="h-4 w-4 text-muted-foreground" />
                )}
              </button>
            </div>
          </div>

          <Button 
            type="submit" 
            className="w-full py-2 sm:py-6 bg-emerald-500 hover:bg-emerald-600 text-white font-medium text-sm sm:text-base"
            disabled={loading}
          >
            {loading ? t("signingIn") : t("signIn")}
            {!loading && <LogIn className="ml-2 h-4 w-4" />}
          </Button>
        </form>
      </div>

      {/* Footer with Social Media Icons */}
      <div className="mt-8 flex items-center justify-center space-x-6">
        <button
          onClick={() => openExternalLink('https://github.com/operacle/checkcle')}
          className="text-muted-foreground hover:text-foreground transition-colors"
          aria-label="GitHub"
        >
          <Github className="h-5 w-5" />
        </button>
        <button
          onClick={() => openExternalLink('https://x.com/checkcle_oss')}
          className="text-muted-foreground hover:text-foreground transition-colors"
          aria-label="X (Twitter)"
        >
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
          </svg>
        </button>
        <button
          onClick={() => openExternalLink('https://discord.gg/xs9gbubGwX')}
          className="text-muted-foreground hover:text-foreground transition-colors"
          aria-label="Discord"
        >
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515a.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0a12.64 12.64 0 0 0-.617-1.25a.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057a19.9 19.9 0 0 0 5.993 3.03a.078.078 0 0 0 .084-.028a14.09 14.09 0 0 0 1.226-1.994a.076.076 0 0 0-.041-.106a13.107 13.107 0 0 1-1.872-.892a.077.077 0 0 1-.008-.128a10.2 10.2 0 0 0 .372-.292a.074.074 0 0 1 .077-.010c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127a12.299 12.299 0 0 1-1.873.892a.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028a19.839 19.839 0 0 0 6.002-3.03a.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419c0-1.333.956-2.419 2.157-2.419c1.21 0 2.176 1.096 2.157 2.42c0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419c0-1.333.955-2.419 2.157-2.419c1.21 0 2.176 1.096 2.157 2.42c0 1.333-.946 2.418-2.157 2.418z"/>
          </svg>
        </button>
      </div>

      <ForgotPasswordDialog 
        open={showForgotPassword} 
        onOpenChange={setShowForgotPassword} 
      />
    </div>
  );
};

export default Login;