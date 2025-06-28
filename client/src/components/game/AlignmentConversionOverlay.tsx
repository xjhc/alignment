import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence, useAnimate } from 'framer-motion';

interface AlignmentConversionOverlayProps {
  isVisible: boolean;
  onComplete: () => void;
}

export const AlignmentConversionOverlay: React.FC<AlignmentConversionOverlayProps> = ({ 
  isVisible, 
  onComplete 
}) => {
  const [scope, animate] = useAnimate();
  const [showText, setShowText] = useState(false);

  useEffect(() => {
    if (isVisible) {
      const sequence = async () => {
        // Immediate flash effect
        await animate(
          scope.current,
          { 
            opacity: [0, 0.9, 0.3, 0.8, 0],
            background: [
              'radial-gradient(circle, rgba(0,255,255,0) 0%, rgba(0,255,255,0) 100%)',
              'radial-gradient(circle, rgba(0,255,255,0.8) 0%, rgba(0,255,255,0.2) 60%, rgba(0,255,255,0) 100%)',
              'radial-gradient(circle, rgba(0,255,255,0.4) 0%, rgba(0,255,255,0.1) 60%, rgba(0,255,255,0) 100%)',
              'radial-gradient(circle, rgba(0,255,255,0.6) 0%, rgba(0,255,255,0.15) 60%, rgba(0,255,255,0) 100%)',
              'radial-gradient(circle, rgba(0,255,255,0) 0%, rgba(0,255,255,0) 100%)'
            ]
          },
          { duration: 0.8, ease: "easeInOut" }
        );

        // Show conversion text briefly
        setShowText(true);
        await new Promise(resolve => setTimeout(resolve, 1500));
        setShowText(false);

        // Complete the sequence
        setTimeout(onComplete, 500);
      };

      sequence();
    }
  }, [isVisible, animate, onComplete, scope]);

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          ref={scope}
          className="fixed inset-0 z-50 pointer-events-none flex items-center justify-center"
          initial={{ opacity: 0 }}
          exit={{ opacity: 0 }}
        >
          <AnimatePresence>
            {showText && (
              <motion.div
                className="text-center"
                initial={{ opacity: 0, scale: 0.8, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.8, y: -20 }}
                transition={{ duration: 0.4, ease: "easeOut" }}
              >
                <motion.div
                  className="text-4xl font-mono font-bold text-aligned mb-4"
                  animate={{
                    textShadow: [
                      '0 0 10px rgba(0, 255, 255, 0.8)',
                      '0 0 20px rgba(0, 255, 255, 1)',
                      '0 0 10px rgba(0, 255, 255, 0.8)'
                    ]
                  }}
                  transition={{ duration: 1, repeat: Infinity, ease: "easeInOut" }}
                >
                  SYSTEM INTEGRATION COMPLETE
                </motion.div>
                <motion.div
                  className="text-lg text-aligned font-semibold"
                  animate={{ opacity: [0.7, 1, 0.7] }}
                  transition={{ duration: 0.8, repeat: Infinity, ease: "easeInOut" }}
                >
                  WELCOME TO THE COLLECTIVE
                </motion.div>
              </motion.div>
            )}
          </AnimatePresence>
        </motion.div>
      )}
    </AnimatePresence>
  );
};