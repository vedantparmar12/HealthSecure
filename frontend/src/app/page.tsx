import Navbar from '@/components/landing/navbar';
import Hero from '@/components/landing/hero';
import Features from '@/components/landing/features';
import Security from '@/components/landing/security';
import Clinicians from '@/components/landing/clinicians';
import CTA from '@/components/landing/cta';
import Footer from '@/components/landing/footer';

export default function HomePage() {
  return (
    <main className="min-h-screen">
      <Navbar />
      <Hero />
      <Features />
      <Security />
      <Clinicians />
      <CTA />
      <Footer />
    </main>
  );
}